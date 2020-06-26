// TODO check for mem leaks
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <time.h>
#include <pthread.h>

#include "reporter/plat/tracking.h"

#ifdef __linux__

static Atom     active_window_prop;
static Display *display;
static Window   root_window;

static bool             tracking_started = false;
static pthread_t        tracking_thread;
static tracking_opt_t  *tracking_opts;

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

static void handle_window_change(Display *display, Window window) {
    static char prog_name[512];

    time_t ts = time(NULL);
    get_current_window_name(display, window, prog_name, 512);
    printf("[%lu] Switched to a new program %s\n", ts, prog_name);
    // TODO Insert report event function call here
    // TODO check if prog_name or window ID changed, no need to report if not
}

static void *event_loop(UNUSED void *arg) {
    debug("Starting tracking with opts: %d|%d|%d (prog|mouse|key)\n",
            tracking_opts->foreground_program, tracking_opts->mouse_click,
            tracking_opts->keystroke);

    long event_mask = 0;
    if (tracking_opts->foreground_program) {
        // receive property change events
        event_mask |= PropertyChangeMask;
        XSelectInput(display, root_window, event_mask);
    }

    XEvent event;
    /* drain any outstanding event if present */
    while (XCheckMaskEvent(display, event_mask, &event));

    while (1) {
        if (!tracking_started) {
            break;
        }
        XMaskEvent(display, event_mask, &event);
        if (event.xproperty.atom != active_window_prop)
            continue;
        handle_window_change(display, root_window);
    }

    // clear event mask on exit of this function
    event_mask = 0;
    XSelectInput(display, root_window, event_mask);

    // TODO clear mouse/keyboard tracking event masks

    return NULL;
}

int init_tracking() {
    if (strcmp(getenv("XDG_SESSION_TYPE"), "x11")) {
        error("Not using x11 as display server, tracking may not be accurate\n");
    }
    XSetErrorHandler(xlib_error_handler);
    // open connection to the X server
    display = XOpenDisplay(NULL);
    if (display == NULL) {
        error("Cannot open connection to X server\n");
        return 1;
    }

    // init Atoms
    active_window_prop = XInternAtom(display, "_NET_ACTIVE_WINDOW", 0);

    // get root window
    root_window = XDefaultRootWindow(display);

    return 0;
}

void exit_tracking() {
    XCloseDisplay(display);
    display = NULL;
}

int start_tracking(tracking_opt_t *opts) {

    if (tracking_started) {
        error("Tracking tracking_started already!\n");
        return 1;
    }

    tracking_opts = opts;

    if (!(opts->foreground_program || opts->mouse_click || opts->keystroke)) {
        debug("Nothing to be tracked, not doing anything\n");
        return 0;
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
    event.type = PropertyNotify;
    event.xproperty.atom = active_window_prop;
    /* manually send a PropertyNotify event to the root window
     * in case the tracking thread is blocked at XMaskEvent */
    if(!XSendEvent(display, root_window, False, PropertyChangeMask, &event))
        error("Failed to send event to root_window "
                "to unblock the tracking thread\n");
    XFlush(display);

    int ret = pthread_join(tracking_thread, NULL);
    if (ret) {
        perror("Cannot join the tracking thread");
        return;
    }


    printf("Tracking stopped\n");
}

#else
#error "This code only works on Linux"
#endif /* #ifdef __linux__ */
