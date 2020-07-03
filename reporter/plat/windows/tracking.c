#include "reporter/plat/tracking.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <time.h>
#include <pthread.h>

#include "reporter/core/core.h"

bool _is_tracking = false;

int init_tracking() {
    return 0;
}

int start_tracking(tracking_opt_t *opts) {
    if (!_is_tracking) {
        _is_tracking = true;
        SendStartTrackingEvent();
    }
}

void stop_tracking() {
    if (_is_tracking) {
        _is_tracking = false;
        SendStopTrackingEvent();
    }
}

bool is_tracking() {
    return _is_tracking;
}
