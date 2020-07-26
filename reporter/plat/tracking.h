#pragma once

#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define UNUSED __attribute__((unused))

#define prod_error(...) fprintf(stderr, __VA_ARGS__)

#ifdef DEBUG
#define prod_debug(...) fprintf(stderr, "[debug] " __VA_ARGS__)
#else
#define prod_debug(...)
#endif

#ifdef __cplusplus
extern "C" {
#endif

struct tracking_option {
  const char *opt_name;
  const char *display_name;
};

extern const struct tracking_option tracking_options[];
extern const size_t NUM_OPTIONS;

int init_tracking();
int start_tracking();
void stop_tracking();
void run_event_loop();
void stop_event_loop();
bool get_option(const char *opt);

#ifdef __cplusplus
}
#endif
