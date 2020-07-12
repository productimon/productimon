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
- (void)init_observers:(tracking_opt_t *)opts {
  if (opts->foreground_program) {
    [[NSWorkspace sharedWorkspace].notificationCenter
        addObserver:self
           selector:@selector(receiveAppSwitchNtfn:)
               name:@"NSWorkspaceDidActivateApplicationNotification"
             object:nil];
  }
}

- (void)remove_observers {
  [[NSWorkspace sharedWorkspace].notificationCenter removeObserver:self];
}

- (void)receiveAppSwitchNtfn:(NSNotification *)notification {
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
