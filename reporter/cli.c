#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "plat/tracking.h"

int main() {
    tracking_opt_t opts = {
        .foreground_program = 1,
        .mouse_click = 1,
        .keystroke = 1,
        .server_addr = "https://api.productimon.com"
    };

    if (init_tracking()) {
        error("Failed to init tracking\n");
        return 1;
    }
    printf("Productimon data reporter CLI\n");
    printf("Valid commands are: start, stop and exit\n");

    char *command = NULL;
    size_t size = 0;
    printf("> ");
    while (getline(&command, &size, stdin) > 0) {
        if (strcmp(command, "start\n") == 0) {
            start_tracking(&opts);
        } else if (strcmp(command, "stop\n") == 0) {
            stop_tracking(&opts);
        } else if (strcmp(command, "exit\n") == 0) {
            break;
        } else {
            printf("unknown command\n");
        }
        printf("> ");
    }
    free(command);
    return 0;
}
