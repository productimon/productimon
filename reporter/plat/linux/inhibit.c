/*
 * Productimon reporter linux lib to stop tracking
 * before system goes into sleep mode
 * and resume tracking after wake up
 */
#include "inhibit.h"

#include <dbus/dbus.h>
#include <pthread.h>
#include <semaphore.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#include "reporter/core/cgo/cgo.h"
#include "reporter/plat/tracking.h"

#define SLEEP_SIGNAL_MATCH                                    \
  "type='signal',interface='org.freedesktop.login1.Manager'," \
  "member='PrepareForSleep'"
#define LOCK_SIGNAL_MATCH                                      \
  "type='signal',interface='org.freedesktop.DBus.Properties'," \
  "member='PropertiesChanged'"

#define DBUS_READ_TIMEOUT (-1)

static inhibit_hook_f sleep_callback;
static inhibit_hook_f wakeup_callback;

static pthread_t inhibit_thread;
static sem_t inhibit_initialised;
static volatile bool inhibit_init_success;
static volatile int inhibit_fd = -1;
static DBusConnection *dbusconn = NULL;

/* Call systemd inhibit through dbus, return the fd or -1 on error */
static int systemd_inhibit(DBusConnection *conn, const char *what,
                           const char *who, const char *why, const char *mode) {
  int fd = 100;
  DBusMessage *msg;
  DBusMessageIter args;
  DBusMessageIter ret_iter;
  DBusError err;
  dbus_error_init(&err);

  /* specify to call systemd-logind's inhibit method */
  msg = dbus_message_new_method_call(
      "org.freedesktop.login1", "/org/freedesktop/login1",
      "org.freedesktop.login1.Manager", "Inhibit");
  if (NULL == msg) {
    prod_error("Message Null\n");
    return -1;
  }

  dbus_message_iter_init_append(msg, &args);
  if (!dbus_message_iter_append_basic(&args, DBUS_TYPE_STRING, &what)) {
    prod_error("No memory for args\n");
    return -1;
  }
  if (!dbus_message_iter_append_basic(&args, DBUS_TYPE_STRING, &who)) {
    prod_error("No memory for args\n");
    return -1;
  }
  if (!dbus_message_iter_append_basic(&args, DBUS_TYPE_STRING, &why)) {
    prod_error("No memory for args\n");
    return -1;
  }
  if (!dbus_message_iter_append_basic(&args, DBUS_TYPE_STRING, &mode)) {
    prod_error("No memory for args\n");
    return -1;
  }

  DBusMessage *reply_msg;
  /* send message and block until reply is available */
  reply_msg = dbus_connection_send_with_reply_and_block(conn, msg, -1, &err);
  if (dbus_error_is_set(&err)) {
    prod_error("Error on sending method call: %s\n", err.message);
    dbus_error_free(&err);
    return -1;
  }
  if (reply_msg == NULL) {
    prod_error("Reply is null\n");
    return -1;
  }

  if (!dbus_message_iter_init(reply_msg, &ret_iter)) {
    prod_error("Failed to init iter for return vals\n");
    return -1;
  }

  dbus_message_iter_get_basic(&ret_iter, &fd);

  /* free message */
  dbus_message_unref(msg);
  dbus_message_unref(reply_msg);
  return fd;
}

static DBusConnection *setup_dbus_conn() {
  DBusConnection *conn;

  DBusError err;
  dbus_error_init(&err);
  conn = dbus_bus_get_private(DBUS_BUS_SYSTEM, &err);
  if (dbus_error_is_set(&err)) {
    prod_error("Error on dbus get: %s\n", err.message);
    dbus_error_free(&err);
    return NULL;
  }
  if (conn == NULL) {
    prod_error("Connection is NULL\n");
    return NULL;
  }

  /* don't call _exit() when it disconnects */
  dbus_connection_set_exit_on_disconnect(conn, false);

  /* subscribe to PrepareForSleep signals */
  dbus_bus_add_match(conn, SLEEP_SIGNAL_MATCH, &err);
  if (dbus_error_is_set(&err)) {
    prod_error("Error on dbus add match: %s\n", err.message);
    return NULL;
  }

  /* subscribe to PropertiesChanged signals */
  dbus_bus_add_match(conn, LOCK_SIGNAL_MATCH, &err);
  if (dbus_error_is_set(&err)) {
    prod_error("Error on dbus add match: %s\n", err.message);
    return NULL;
  }

  dbus_connection_flush(conn);

  /* drain any incoming signals */
  while (dbus_connection_read_write(conn, 0)) {
    DBusMessage *msg = dbus_connection_pop_message(conn);
    if (msg == NULL)
      break;
    else
      dbus_message_unref(msg);
  }
  return conn;
}

static bool handle_sleep_message(DBusMessage *msg, dbus_int32_t *sleeping) {
  DBusMessageIter args;
  if (!dbus_message_iter_init(msg, &args)) return false;
  if (dbus_message_iter_get_arg_type(&args) != DBUS_TYPE_BOOLEAN) return false;
  dbus_message_iter_get_basic(&args, sleeping);
  prod_debug("Got PrepareForSleep signal with value %d\n", *sleeping);
  return true;
}

static bool handle_properties_changed_message(DBusMessage *msg,
                                              dbus_int32_t *sleeping) {
  DBusMessageIter args, array_iter, dict_entry, prop_val;
  char *attribute;
  if (!dbus_message_iter_init(msg, &args)) return false;
  if (dbus_message_iter_get_arg_type(&args) != DBUS_TYPE_STRING) return false;
  dbus_message_iter_get_basic(&args, &attribute);
  if (strcmp(attribute, "org.freedesktop.login1.Session") != 0) return false;
  if (!dbus_message_iter_next(&args)) return false;
  if (dbus_message_iter_get_arg_type(&args) != DBUS_TYPE_ARRAY) return false;
  dbus_message_iter_recurse(&args, &array_iter);
  do {
    if (dbus_message_iter_get_arg_type(&array_iter) != DBUS_TYPE_DICT_ENTRY)
      return false;
    dbus_message_iter_recurse(&array_iter, &dict_entry);
    if (dbus_message_iter_get_arg_type(&dict_entry) != DBUS_TYPE_STRING)
      return false;
    dbus_message_iter_get_basic(&dict_entry, &attribute);
    if (strcmp(attribute, "LockedHint") == 0) {
      if (!dbus_message_iter_next(&dict_entry)) return false;
      if (dbus_message_iter_get_arg_type(&dict_entry) != DBUS_TYPE_VARIANT)
        return false;
      dbus_message_iter_recurse(&dict_entry, &prop_val);
      if (dbus_message_iter_get_arg_type(&prop_val) != DBUS_TYPE_BOOLEAN)
        return false;
      dbus_message_iter_get_basic(&prop_val, sleeping);
      prod_debug("Got PropertiesChanged signal with value %d\n", *sleeping);
      return true;
    }
  } while (dbus_message_iter_next(&array_iter));
  return false;
}

/* runs in a separate thread to receive msg from dbus */
void *dbus_receive_msg_loop(UNUSED void *unused) {
  dbusconn = setup_dbus_conn();
  if (dbusconn == NULL) {
    goto exit;
  }

  /* register inhibitor */
  inhibit_fd = systemd_inhibit(dbusconn, "sleep", "Productimon",
                               "Stop stracking...", "delay");
  if (inhibit_fd < 0) {
    goto cleanup;
  }
  prod_debug("got inhibit fd %d\n", inhibit_fd);

  /* init complete */
  inhibit_init_success = true;
  sem_post(&inhibit_initialised);

  DBusMessage *msg;
  dbus_int32_t sleeping = false;

  while (1) {
    /* block until there's incoming message */
    if (!dbus_connection_read_write(dbusconn, DBUS_READ_TIMEOUT)) {
      prod_error("dbus connection broke\n");
      break;
    }

    while ((msg = dbus_connection_pop_message(dbusconn)) != NULL) {
      bool trigger = false;
      if (dbus_message_is_signal(msg, "org.freedesktop.login1.Manager",
                                 "PrepareForSleep")) {
        trigger = handle_sleep_message(msg, &sleeping);
      } else if (dbus_message_is_signal(msg, "org.freedesktop.DBus.Properties",
                                        "PropertiesChanged")) {
        trigger = handle_properties_changed_message(msg, &sleeping);
      }
      if (trigger) {
        if (sleeping) {
          sleep_callback();
          if (inhibit_fd != -1) close(inhibit_fd);
          inhibit_fd = -1;
        } else {
          wakeup_callback();
          inhibit_fd = systemd_inhibit(dbusconn, "sleep", "Productimon",
                                       "Stop stracking...", "delay");
          prod_debug("Got new inhibit fd %d\n", inhibit_fd);
        }
      }
      /* free msg */
      dbus_message_unref(msg);
    }
  }
cleanup:
  /* it's safe to close multiple times but we must close before unref */
  dbus_connection_close(dbusconn);
  dbus_connection_unref(dbusconn);
  dbusconn = NULL;

exit:
  return NULL;
}

int init_inhibit(inhibit_hook_f _sleep_callback,
                 inhibit_hook_f _wakeup_callback) {
  /* tell D-Bus to do locking */
  /* since D-Bus 1.7 it is safe to call this function from any thread, any
   * number of times */
  /* this is the first entry-point in inhibit.c */
  dbus_threads_init_default();
  sleep_callback = _sleep_callback;
  wakeup_callback = _wakeup_callback;
  if (sem_init(&inhibit_initialised, 0, 0)) {
    perror("Error to init sem");
    return 1;
  }
  if (pthread_create(&inhibit_thread, NULL, dbus_receive_msg_loop, NULL)) {
    prod_error("Failed to create dbus thread\n");
    return 1;
  }

  sem_wait(&inhibit_initialised);
  if (!inhibit_init_success) {
    prod_error("inhibit init failed\n");
    return 1;
  }
  prod_debug("inhibit init sucess\n");
  return 0;
}

void wait_inhibit_cleanup() {
  if (inhibit_fd != -1) {
    close(inhibit_fd);
    inhibit_fd = -1;
  }
  if (dbusconn != NULL) {
    dbus_connection_close(dbusconn);
  }
  pthread_join(inhibit_thread, NULL);
  prod_debug("inhibit thread exit\n");
}
