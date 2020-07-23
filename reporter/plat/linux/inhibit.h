#pragma once

#include <stdbool.h>

typedef void (*inhibit_hook_f)();

int init_inhibit(inhibit_hook_f _sleep_callback,
                 inhibit_hook_f _wakeup_callback);
void wait_inhibit_cleanup();
