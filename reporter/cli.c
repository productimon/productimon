#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "reporter/core/core.h"
#include "reporter/plat/tracking.h"

#define MAX_CMD_LEN BUFSIZ

static char command[MAX_CMD_LEN];

int main(int argc, const char* argv[]) {
    setbuf(stdout, NULL);
    setbuf(stderr, NULL);
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

    printf("> ");
    while (fgets(command, MAX_CMD_LEN, stdin) != NULL) {
        if (strcmp(command, "start\n") == 0) {
            start_tracking(&opts);
        } else if (strcmp(command, "stop\n") == 0) {
            stop_tracking(&opts);
        } else if (strcmp(command, "exit\n") == 0) {
            break;
        } /* else {
            // NOTE this breaks the offline buffering feature since message queue in the core
            // is not a buffered channel and this could be blocking
            // TODO delete this and relevant code in the core
            SendMessage(command);  // demo sending something to core
            printf("unknown command\n");
        } */
        printf("> ");
    }

    QuitReporter(is_tracking());
    return 0;
}
