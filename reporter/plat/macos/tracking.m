#include "reporter/plat/tracking.h"

#import <AppKit/AppKit.h>
#include <stdbool.h>
#include <stdio.h>

#include "reporter/core/core.h"

@interface Tracking : NSObject
- (void)init_observers:(tracking_opt_t *)opts;
- (void)remove_observers;
@end

@implementation Tracking
id event_handler;

- (void)init_observers:(tracking_opt_t *)opts {
  if (opts->foreground_program) {
    [[NSWorkspace sharedWorkspace].notificationCenter
        addObserver:self
           selector:@selector(app_switch_handler:)
               name:@"NSWorkspaceDidActivateApplicationNotification"
             object:nil];
  }

  NSEventMask mask = 0;
  if (opts->keystroke) mask |= NSEventMaskKeyDown;
  if (opts->keystroke)
    mask |= NSEventMaskLeftMouseDown | NSEventMaskRightMouseDown | NSEventMaskScrollWheel;
  event_handler =
      [NSEvent addGlobalMonitorForEventsMatchingMask:mask
                                             handler:^(NSEvent *event) {
                                               switch (event.type) {
                                                 case NSEventTypeLeftMouseDown:
                                                 case NSEventTypeRightMouseDown:
                                                 case NSEventTypeScrollWheel:
                                                   HandleMouseClick();
                                                   break;
                                                 case NSEventTypeKeyDown:
                                                   HandleKeystroke();
                                                   break;
                                                 default:
                                                   NSLog(@"Unexpected event %@\n", event);
                                                   break;
                                               }
                                             }];
  NSLog(@"Got event handler %@\n", event_handler);
}

- (void)remove_observers {
  [[NSWorkspace sharedWorkspace].notificationCenter removeObserver:self];
  [NSEvent removeMonitor:event_handler];
}

- (void)app_switch_handler:(NSNotification *)notification {
  NSRunningApplication *app = notification.userInfo[@"NSWorkspaceApplicationKey"];
  const char *app_name = [app.localizedName UTF8String];
  debug("Switched to %s\n", app_name);
  SendWindowSwitchEvent((char *)app_name);
}
@end

static tracking_opt_t *tracking_opts = NULL;
static bool tracking_started = false;
static Tracking *tracking;

void run_event_loop() {
  tracking = [Tracking new];
  [NSApplication sharedApplication];
  [NSApp run];
}

void stop_event_loop() { [NSApp terminate:nil]; }

int init_tracking() { return 0; }

int start_tracking(tracking_opt_t *opts) {
  if (tracking_started) {
    error("Tracking started already!\n");
    return 1;
  }
  tracking_opts = opts;

  SendStartTrackingEvent();
  [tracking init_observers:tracking_opts];
  debug("Tracking started\n");

  tracking_started = true;
  return 0;
}

void stop_tracking() {
  if (!tracking_started) {
    error("Tracking stopped already!\n");
    return;
  }

  [tracking remove_observers];
  SendStopTrackingEvent();
  debug("Tracking stopped\n");
  tracking_started = false;
}

bool is_tracking() { return tracking_started; }
