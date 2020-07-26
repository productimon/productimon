#include "reporter/plat/tracking.h"

#include <stdbool.h>
#include <stdlib.h>
#include <string.h>

#include "reporter/core/cgo/cgo.h"

const struct tracking_option tracking_options[] = {
    {.opt_name = "autorun", .display_name = "Auto run at startup"},
    {.opt_name = "foreground_program", .display_name = "Foreground Programs"},
    {.opt_name = "mouse_click", .display_name = "Mouse Click Statistics"},
    {.opt_name = "keystroke", .display_name = "Keystroke Statistics"},
};
const size_t NUM_OPTIONS = sizeof(tracking_options) / sizeof(*tracking_options);

bool get_option(const char *opt) {
  return ProdCoreIsOptionEnabled((char *)opt);
}
