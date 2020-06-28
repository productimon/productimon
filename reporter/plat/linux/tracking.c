// TODO check for mem leaks
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/XKBlib.h>
#include <X11/extensions/XInput2.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <time.h>
#include <pthread.h>

#include "reporter/core/core.h"
#include "reporter/plat/tracking.h"

#ifdef __linux__

static Atom     active_window_prop;
static Display *display;
static Window   root_window;

static volatile bool    tracking_started = false;
static pthread_t        tracking_thread;
static tracking_opt_t  *tracking_opts;

static int              xi_major_opcode;
static int              tracking_n_clicks;
static int              tracking_n_keystrokes;

static int xlib_error_handler(Display *display, XErrorEvent *event) {
    char buf[256];
    XGetErrorText(display, event->type, buf, 256);
    error("xlib error: %s\n", buf);
    XCloseDisplay(display);
    exit(1);
}

static int get_current_window_name(Display *display, Window root,
        char *buf, int size) {
    Atom actual_type_ret;
    int actual_format_ret;
    unsigned long nitems_return;
    unsigned long bytes_after_return;
    unsigned char *prop_return;

    /* Get current active window's ID */
    int ret = XGetWindowProperty(display, root, active_window_prop,
            0, 4, 0, AnyPropertyType, &actual_type_ret,
            &actual_format_ret, &nitems_return, &bytes_after_return,
            &prop_return);

    if (ret || actual_format_ret != 32 || nitems_return != 1) {
        debug("atom returned %s\n", XGetAtomName(display, actual_type_ret));
        error("Failed to get active window\n");
        return -1;
    }
    /* debug("fmt ret: %d, %lu items, %lu bytes remains, prop @ %p\n", */
    /*         actual_format_ret, nitems_return, bytes_after_return,
     *         prop_return); */

    Window active_window = *(Window *) prop_return;
    debug("Got active window ID 0x%lX\n", active_window);
    XFree(prop_return);

    if (active_window == 0) {
        strncpy(buf, "Desktop", size);
        return 0;
    }

    /* Get its WM_CLASS class name */
    XClassHint class_hint;
    XGetClassHint(display, active_window, &class_hint);
    debug("WM_CLASS: %s\n", class_hint.res_class);
    snprintf(buf, size, "%s", class_hint.res_class);
    XFree(class_hint.res_name);
    XFree(class_hint.res_class);
    return 0;
}

static int send_input_stats() {
    printf("%d keystrokes in last window session\n", tracking_n_keystrokes);
    printf("%d mouse clicks in last window session\n", tracking_n_clicks);
    // TODO insert acutal code to send stats to server
    return 1;
}

static void handle_window_change(Display *display, Window window) {
    static char prog_name[512] = {0};
    char new_prog_name[512];

    get_current_window_name(display, window, new_prog_name, 512);

    /* no need to report anything if still using the same program */
    if (!strcmp(new_prog_name, prog_name))
        return;

    strcpy(prog_name, new_prog_name);

    send_input_stats();
    SendWindowSwitchEvent(prog_name);
}

// TODO: these event handlers can be defined in the go library
// and we'll call them from within the plat code
// the go lib will take care of maintain and regularly send input stats
static void handle_key_press() {
    tracking_n_keystrokes++;
    debug("key press detected\n");
}

static void handle_button_press() {
    tracking_n_clicks++;
    debug("button press detected\n");
}

static int check_x_input_lib(Display *display) {
    int unused1, unused2;
    xi_major_opcode = 0;
    if (!XQueryExtension(display, "XInputExtension",
                &xi_major_opcode, &unused1, &unused2)) {
        error("X Input extension not available\n");
        return 1;
    }
    /* request XI 2.0 */
    int major = 2, minor = 0;
    int queryResult = XIQueryVersion(display, &major, &minor);
    if (queryResult == BadRequest) {
        error("Need X Input 2.0 (got %d.%d)\n", major, minor);
        return 1;
    } else if (queryResult != Success) {
        error("XIQueryVersion failed\n");
        return 1;
    }
    debug("X Input Extension (%d.%d)\n", major, minor);
    return 0;
}

static void *event_loop(UNUSED void *arg) {
    debug("Starting tracking with opts: %d|%d|%d (prog|mouse|key)\n",
            tracking_opts->foreground_program, tracking_opts->mouse_click,
            tracking_opts->keystroke);

    SendStartTrackingEvent();

    long event_mask = 0;
    if (tracking_opts->foreground_program) {
        // receive property change events
        event_mask |= PropertyChangeMask;
        XSelectInput(display, root_window, event_mask);
    }

    XIEventMask xi_event_mask;
    if (tracking_opts->keystroke || tracking_opts->mouse_click) {
        tracking_n_clicks = 0;
        tracking_n_keystrokes = 0;
        unsigned char xi_mask_val[(XI_LASTEVENT + 7 / 8)] = {0};
        xi_event_mask.deviceid = XIAllMasterDevices;
        xi_event_mask.mask_len = sizeof(xi_mask_val);
        xi_event_mask.mask = xi_mask_val;
        XISetMask(xi_mask_val, XI_RawKeyPress);
        XISetMask(xi_mask_val, XI_RawButtonPress);
        XISelectEvents(display, root_window, &xi_event_mask, 1);
        /* XSync(display, false); */
    }

    XEvent event;
    XGenericEventCookie *cookie = (XGenericEventCookie*)&event.xcookie;

    while (1) {
        if (!tracking_started) {
            break;
        }
        XNextEvent(display, &event);
        if (event.type == PropertyNotify &&
                event.xproperty.atom == active_window_prop)
            handle_window_change(display, root_window);
        if (XGetEventData(display, cookie) &&
                cookie->type == GenericEvent &&
                cookie->extension == xi_major_opcode) {
            if (cookie->evtype == XI_RawKeyPress) {
                handle_key_press();
            } else if (cookie->evtype == XI_RawButtonPress) {
                handle_button_press();
            }
        }
    }

    SendStopTrackingEvent();
    return NULL;
}

int init_tracking() {
    if (strcmp(getenv("XDG_SESSION_TYPE"), "x11")) {
        error("Not using x11 as display server, tracking may not be accurate\n");
    }
    XSetErrorHandler(xlib_error_handler);
    return 0;
}

void exit_tracking() {
}

int start_tracking(tracking_opt_t *opts) {
    // open connection to the X server
    display = XOpenDisplay(NULL);
    if (display == NULL) {
        error("Cannot open connection to X server\n");
        return 1;
    }

    if (check_x_input_lib(display)) {
        xi_major_opcode = 0;
        error("X Input not available, no mouse/key tracking\n");
    }

    // init Atoms
    active_window_prop = XInternAtom(display, "_NET_ACTIVE_WINDOW", 0);

    // get root window
    root_window = XDefaultRootWindow(display);


    if (tracking_started) {
        error("Tracking tracking_started already!\n");
        return 1;
    }

    tracking_opts = opts;

    if (!(opts->foreground_program || opts->mouse_click || opts->keystroke)) {
        debug("Nothing to be tracked, not doing anything\n");
        return 0;
    }


    if ((opts->mouse_click || opts->keystroke) && !xi_major_opcode) {
        error("Requested mouse/key stats but X input lib not available!\n");
        opts->mouse_click = false;
        opts->keystroke = false;
    }

    int ret = pthread_create(&tracking_thread, NULL, event_loop, display);
    if (ret) {
        perror("Cannot create the tracking thread");
        return 1;
    }
    tracking_started = true;
    printf("Tracking tracking_started\n");
    return 0;
}

void stop_tracking() {
    if (!tracking_started) {
        error("Tracking not started, not doing anything\n");
        return;
    }
    debug("Stopping tracking\n");

    tracking_started = false;

    XEvent event;
    event.type = ClientMessage;
    /* length of data payload in the event, need this otherwise xlib will complain */
    event.xclient.format = 32;
    /* manually send a ClientMessage event to the root window
     * in case the tracking thread is blocked at XNextEvent */
    if(!XSendEvent(display, root_window, False, PropertyChangeMask, &event))
        error("Failed to send event to root_window "
                "to unblock the tracking thread\n");
    XFlush(display);

    // TODO I think I got a race somehow and this deadlocked
    // think about it and fix it
    int ret = pthread_join(tracking_thread, NULL);
    if (ret) {
        perror("Cannot join the tracking thread");
        return;
    }

    XCloseDisplay(display);
    display = NULL;
    printf("Tracking stopped\n");
}

#else
#error "This code only works on Linux"
#endif /* #ifdef __linux__ */
