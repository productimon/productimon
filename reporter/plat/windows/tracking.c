#include "reporter/plat/tracking.h"

#include <processthreadsapi.h>
#include <psapi.h>
#include <shlwapi.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <windows.h>
#include <winuser.h>
#include <winver.h>

#include "reporter/core/core.h"

#define STOP_MSG (WM_USER + 1)
#define STOP_W_PARAM ((WPARAM)0xDEADBEEF)
#define STOP_L_PARAM ((LPARAM)0xBADDCAFE)
#define IS_STOP_MSG(msg)                                    \
  (msg.message == STOP_MSG && msg.wParam == STOP_W_PARAM && \
   msg.lParam == STOP_L_PARAM)

static bool tracking_started = false;
static HANDLE tracking_thread = NULL;
static DWORD tracking_thread_id;
static tracking_opt_t *tracking_opts;

static int get_name_from_handle(HWND hwnd, char *buf, size_t size) {
  int ret = 1;
  DWORD pid;
  GetWindowThreadProcessId(hwnd, &pid);
  debug("got pid %d\n", pid);

  HANDLE proc_handle = OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, 0, pid);
  if (proc_handle == NULL) {
    error("OpenProcess failed\n");
    return 1;
  }

  char path[MAX_PATH];
  if (GetModuleFileNameExA(proc_handle, NULL, path, MAX_PATH) == 0) {
    error("Failed to get executable path\n");
    goto error_close_handle;
  }
  debug("exec full path: %s\n", path);

  DWORD ver_info_size = GetFileVersionInfoSizeA(path, NULL);
  if (ver_info_size == 0) {
    debug("Failed to get version info size\n");
    PathStripPathA(path);
    snprintf(buf, size, "%s", path);
    CloseHandle(proc_handle);
    return 0;
  }
  debug("version info size: %ld\n", ver_info_size);

  void *version_buf = malloc(ver_info_size);
  if (version_buf == NULL) {
    error("Failed to allocate buffer\n");
    goto error_close_handle;
  }

  if (!GetFileVersionInfoA(path, 0, ver_info_size, version_buf)) {
    error("Failed to get version info\n");
    goto error_free_version_buf;
  }

  /* Get all version info translations */
  static struct LANGANDCODEPAGE {
    WORD wLanguage;
    WORD wCodePage;
  } * translate;

  int translate_size;
  VerQueryValueA(version_buf, TEXT("\\VarFileInfo\\Translation"),
                 (LPVOID *)&translate, &translate_size);

  int n_translations = translate_size / sizeof(struct LANGANDCODEPAGE);
  debug("Got %d translations\n", n_translations);

  // NOTE: it seems like all app on my windows have one translation
  if (n_translations < 1) {
    debug("Failed to get any version translations, using exec name instead\n");
    goto use_exec_name;
  }

  char query_str[64];
  snprintf(query_str, 64, "\\StringFileInfo\\%04x%04x\\FileDescription",
           translate[0].wLanguage, translate[0].wCodePage);
  debug("using query: %s\n", query_str);

  char *file_description;
  /* Get program description from the version info */
  if (!VerQueryValueA(version_buf, query_str, (LPVOID *)&file_description,
                      NULL)) {
    debug(
        "Failed to get description from version info, use exec name instead\n");
    goto use_exec_name;
  }
  snprintf(buf, size, "%s", file_description);
  goto success;

use_exec_name:
  PathStripPathA(path);
  snprintf(buf, size, "%s", path);
success:
  ret = 0;

error_free_version_buf:
  free(version_buf);
error_close_handle:
  CloseHandle(proc_handle);
  return ret;
}

static VOID CALLBACK callback(HWINEVENTHOOK hWinEventHook, DWORD dwEvent,
                              HWND hwnd, LONG idObject, LONG idChild,
                              DWORD dwEventThread, DWORD dwmsEventTime) {
  debug("Callback: event %ld, hwnd %d, idObject %ld, idChild %ld, time %ld\n",
        dwEvent, hwnd, idObject, idChild, dwmsEventTime);

  static char prog_name[512] = {0};
  char new_prog_name[512];
  if (get_name_from_handle(hwnd, new_prog_name, 512)) {
    // TODO handle error here
    // we either stop tracking or send an event to switch to an "unknwon"
    // program
    // otherwise viewer will think the user was using the old program all
    // the time...
    // same on linux
    error("Failed to get a name for new window\n");
    return;
  }
  debug("=======> %s <=======\n", new_prog_name);
  if (strcmp(prog_name, new_prog_name) == 0) {
    debug("Switch event triggered but program name is the same\n");
    return;
  }
  strcpy(prog_name, new_prog_name);
  printf("Got new program: %s\n", prog_name);
  SendWindowSwitchEvent(prog_name);
}

DWORD WINAPI tracking_loop(_In_ LPVOID lpParameter) {
  SendStartTrackingEvent();
  debug("Tracking started\n");

  // https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setwineventhook?redirectedfrom=MSDN
  if (tracking_opts->foreground_program) {
    HWINEVENTHOOK event_hook = SetWinEventHook(
        EVENT_SYSTEM_FOREGROUND, EVENT_SYSTEM_FOREGROUND, NULL, callback, 0, 0,
        WINEVENT_OUTOFCONTEXT | WINEVENT_SKIPOWNPROCESS);
    debug("SetWinEventHook got %d\n", event_hook);
  }

  MSG msg;
  while (GetMessage(&msg, NULL, 0, 0)) {
    if (IS_STOP_MSG(msg)) break;

    TranslateMessage(&msg);
    DispatchMessage(&msg);
  }
  SendStopTrackingEvent();
  debug("Tracking stopped\n");
  return 0;
}

int init_tracking() { return 0; }

int start_tracking(tracking_opt_t *opts) {
  if (tracking_started) {
    error("tracking started already!\n");
    return 1;
  }

  tracking_opts = opts;
  if (!(opts->foreground_program || opts->mouse_click || opts->keystroke)) {
    debug("Nothing to be tracked, not doing anything\n");
    return 1;
  }

  tracking_thread =
      CreateThread(NULL, 0, tracking_loop, NULL, 0, &tracking_thread_id);
  if (tracking_thread == NULL) {
    error("Failed to create tracking thread\n");
    return 1;
  }

  tracking_started = true;
  return 0;
}

void stop_tracking() {
  if (!tracking_started) {
    error("tracking stopped already!\n");
    return;
  }

  if (!PostThreadMessageA(tracking_thread_id, STOP_MSG, STOP_W_PARAM,
                          STOP_L_PARAM))
    error("Failed to send stop message: %lu\n", GetLastError());

  WaitForSingleObject(tracking_thread, INFINITE);
  tracking_started = false;
  return;
}

bool is_tracking() { return tracking_started; }
