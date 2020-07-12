#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "reporter/core/core.h"
#include "reporter/plat/tracking.h"

#define MAX_CMD_LEN BUFSIZ

static char command[MAX_CMD_LEN];

void *command_loop(UNUSED void *arg) {
  // TODO: use core module config for this
  tracking_opt_t opts = {
      .foreground_program = 1, .mouse_click = 1, .keystroke = 1};

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
    } else {
      printf("unknown command\n");
    }
    printf("> ");
  }
  QuitReporter(is_tracking());
  stop_event_loop();
  return NULL;
}

int main(int argc, const char *argv[]) {
  pthread_t cli_thread;
  setbuf(stdout, NULL);
  setbuf(stderr, NULL);
  ReadConfig();

  if (!InitReporter()) {
    error("Failed to init core module\n");
    return 1;
  }
  if (init_tracking()) {
    error("Failed to init tracking\n");
    return 1;
  }

  pthread_create(&cli_thread, NULL, command_loop, NULL);
  run_event_loop();
  return 0;
}
