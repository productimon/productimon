#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "reporter/core/core.h"
#include "reporter/plat/tracking.h"

int main(int argc, const char* argv[]) {
    if (argc != 4) {
        error("Usage: %s server username password\n", argv[0]);
        return 1;
    }
    tracking_opt_t opts = {
        .foreground_program = 1, .mouse_click = 1, .keystroke = 1};
    // TODO: use a configuration file instead of cli argument
    if (!InitReporter(argv[1], argv[2], argv[3])) {
        error("Failed to init core module\n");
        return 1;
    }
    if (init_tracking()) {
        error("Failed to init tracking\n");
        return 1;
    }
    printf("Productimon data reporter CLI\n");
    printf("Valid commands are: start, stop and exit\n");

    char* command = NULL;
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
            SendMessage(command);  // demo sending something to core
            printf("unknown command\n");
        }
        printf("> ");
    }
    free(command);
    return 0;
}
