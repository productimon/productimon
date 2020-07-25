#include "reporter/plat/tracking.h"

#define UNICODE

#include <processthreadsapi.h>
#include <psapi.h>
#include <shlwapi.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <strsafe.h>
#include <windows.h>
#include <winnt.h>
#include <winuser.h>
#include <winver.h>
#include <wtsapi32.h>

#include "reporter/core/cgo/cgo.h"

#define STOP_MSG (WM_USER + 1)
#define STOP_W_PARAM ((WPARAM)0xDEADBEEF)
#define STOP_L_PARAM ((LPARAM)0xBADDCAFE)
#define IS_STOP_MSG(msg)                                    \
  (msg.message == STOP_MSG && msg.wParam == STOP_W_PARAM && \
   msg.lParam == STOP_L_PARAM)

static HANDLE tracking_mutex = NULL;
static HANDLE tracking_thread = NULL;
static DWORD tracking_thread_id;
static tracking_opt_t *tracking_opts;

static HWND window_handle;
static HHOOK mouse_hook;
static HHOOK key_hook;
static HWINEVENTHOOK window_change_hook;

static HANDLE event_loop_finished;

static int get_name_from_handle(HWND hwnd, char *buf, size_t size) {
  int ret = 1;
  DWORD pid;
  GetWindowThreadProcessId(hwnd, &pid);

  HANDLE proc_handle = OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, 0, pid);
  if (proc_handle == NULL) {
    prod_error("OpenProcess failed: %d\n", GetLastError());
    return 1;
  }

  wchar_t path[MAX_PATH];
  if (GetModuleFileNameExW(proc_handle, NULL, path, MAX_PATH) == 0) {
    prod_error("Failed to get executable path %d\n", GetLastError);
    goto error_close_handle;
  }

  DWORD ver_info_size = GetFileVersionInfoSizeW(path, NULL);
  if (ver_info_size == 0) {
    prod_debug("Failed to get version info size\n");
    CloseHandle(proc_handle);
    PathStripPathW(path);
    if (!WideCharToMultiByte(CP_UTF8, 0, path, -1, buf, size, NULL, NULL)) {
      prod_error("Failed to convert encoding: %d\n", GetLastError());
      return 1;
    }
    return 0;
  }

  void *version_buf = malloc(ver_info_size);
  if (version_buf == NULL) {
    prod_error("Failed to allocate buffer\n");
    goto error_close_handle;
  }

  if (!GetFileVersionInfoW(path, 0, ver_info_size, version_buf)) {
    prod_error("Failed to get version info\n");
    goto error_free_version_buf;
  }

  /* Get all version info translations */
  static struct LANGANDCODEPAGE {
    WORD wLanguage;
    WORD wCodePage;
  } * translate;

  int translate_size;
  VerQueryValueW(version_buf, TEXT("\\VarFileInfo\\Translation"),
                 (LPVOID *)&translate, &translate_size);

  int n_translations = translate_size / sizeof(struct LANGANDCODEPAGE);

  // NOTE: it seems like all app on my windows have one translation
  if (n_translations < 1) {
    prod_debug(
        "Failed to get any version translations, using exec name instead\n");
    goto use_exec_name;
  }

  wchar_t query_str[64];
  StringCchPrintfW(query_str, 64,
                   TEXT("\\StringFileInfo\\%04x%04x\\FileDescription"),
                   translate[0].wLanguage, translate[0].wCodePage);

  wchar_t *file_description;
  /* Get program description from the version info */
  if (!VerQueryValueW(version_buf, query_str, (LPVOID *)&file_description,
                      NULL)) {
    prod_debug(
        "Failed to get description from version info, use exec name instead\n");
    goto use_exec_name;
  }
  if (!WideCharToMultiByte(CP_UTF8, 0, file_description, -1, buf, size, NULL,
                           NULL)) {
    prod_error("Failed to convert encoding: %d\n", GetLastError());
    goto use_exec_name;
  }
  goto success;

use_exec_name:
  PathStripPathW(path);
  if (!WideCharToMultiByte(CP_UTF8, 0, path, -1, buf, size, NULL, NULL)) {
    prod_error("Failed to convert encoding: %d\n", GetLastError());
    goto error_free_version_buf;
  }
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
  prod_debug(
      "Callback: event %ld, hwnd %d, idObject %ld, idChild %ld, time %ld\n",
      dwEvent, hwnd, idObject, idChild, dwmsEventTime);

  char prog_name[512];
  if (get_name_from_handle(hwnd, prog_name, 512)) {
    prod_error("Failed to get a name for new window\n");
    prog_name[0] = '\0';  // core will set it to Unknown
  }
  prod_debug("Got new program: %s\n", prog_name);
  ProdCoreSwitchWindow(prog_name);
}
static LRESULT CALLBACK keystroke_callback(_In_ int nCode, _In_ WPARAM wParam,
                                           _In_ LPARAM lParam) {
  /* following what the documentation says */
  if (nCode < 0) return CallNextHookEx(NULL, nCode, wParam, lParam);

  /* only do reporting for key down event */
  if (wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN) {
    ProdCoreHandleKeystroke();
  }
  return CallNextHookEx(NULL, nCode, wParam, lParam);
}

static LRESULT CALLBACK mouseclick_callback(_In_ int nCode, _In_ WPARAM wParam,
                                            _In_ LPARAM lParam) {
  /* following what the documentation says */
  if (nCode < 0) return CallNextHookEx(NULL, nCode, wParam, lParam);

  /* only do reporting for key down event */
  if (wParam == WM_LBUTTONDOWN || wParam == WM_RBUTTONDOWN ||
      wParam == WM_MOUSEWHEEL) {
    ProdCoreHandleMouseClick();
  }
  return CallNextHookEx(NULL, nCode, wParam, lParam);
}

static int install_hooks(bool register_session_ntfn) {
  if (tracking_opts->foreground_program) {
    // https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setwineventhook?redirectedfrom=MSDN
    window_change_hook = SetWinEventHook(
        EVENT_SYSTEM_FOREGROUND, EVENT_SYSTEM_FOREGROUND, NULL, callback, 0, 0,
        WINEVENT_OUTOFCONTEXT | WINEVENT_SKIPOWNPROCESS);
    prod_debug("SetWinEventHook got %d\n", window_change_hook);
    if (window_change_hook == NULL) {
      prod_error("Failed to set hook for keyboard: %d\n", GetLastError());
      return 1;
    }
  }

  if (tracking_opts->keystroke) {
    key_hook = SetWindowsHookExA(WH_KEYBOARD_LL, keystroke_callback, NULL, 0);
    prod_debug("SetWindowsHookExA for keyboard got %d\n", key_hook);
    if (key_hook == NULL) {
      prod_error("Failed to set hook for keyboard: %d\n", GetLastError());
      return 1;
    }
  }

  if (tracking_opts->mouse_click) {
    mouse_hook = SetWindowsHookExA(WH_MOUSE_LL, mouseclick_callback, NULL, 0);
    prod_debug("SetWindowsHookExA for mouseclick got %d\n", mouse_hook);
    if (mouse_hook == NULL) {
      prod_error("Failed to set hook for mouseclick: %d\n", GetLastError());
      return 1;
    }
  }

  /* register for lock/unlock and login/logout events
   * doing this twice within the same thread can undo the effect
   * thus the boolean param
   */
  if (register_session_ntfn &&
      !WTSRegisterSessionNotification(window_handle, NOTIFY_FOR_THIS_SESSION)) {
    prod_error("Failed to regitster for session change notifications: %d\n",
               GetLastError());
    return 1;
  }
  return 0;
}

static void suspend_tracking() {
  WaitForSingleObject(tracking_mutex, INFINITE);
  if (tracking_opts->foreground_program && !UnhookWinEvent(window_change_hook))
    prod_error("Failed to remove window change hook: %d\n", GetLastError());

  if (tracking_opts->keystroke && !UnhookWindowsHookEx(key_hook))
    prod_error("Failed to remove key hook: %d\n", GetLastError());

  if (tracking_opts->mouse_click && !UnhookWindowsHookEx(mouse_hook))
    prod_error("Failed to remove mosue hook: %d\n", GetLastError());

  ProdCoreStopTracking();

  ReleaseMutex(tracking_mutex);
}

static void resume_tracking() {
  WaitForSingleObject(tracking_mutex, INFINITE);
  ProdCoreStartTracking();

  install_hooks(false);

  ReleaseMutex(tracking_mutex);
}

static LRESULT CALLBACK session_change_callback(WPARAM type) {
  switch (type) {
    case WTS_SESSION_LOCK:
      prod_debug("system about to lock, suspend tracking...\n");
      suspend_tracking();
      break;
    case WTS_SESSION_UNLOCK:
      prod_debug("login detected, resume tracking\n");
      resume_tracking();
      break;
    case WTS_SESSION_LOGOFF:
      // TODO
      break;
  }
}

static DWORD WINAPI tracking_loop(_In_ LPVOID lpParameter) {
  ProdCoreStartTracking();
  prod_debug("Tracking started\n");

  /* create a hidden message window */
  window_handle = CreateWindowExA(WS_EX_ACCEPTFILES, "Button", "null", 0, 0, 0,
                                  0, 0, HWND_MESSAGE, NULL, NULL, NULL);
  if (window_handle == NULL) {
    prod_error("Failed to create a message window: %d\n", GetLastError());
    return 0;  // use synchronisation primitives here to notify the failure to
               // start_tracking
  }

  if (install_hooks(true)) {
    return 1;  // TODO use sync primitives to have start_tracking wait for this
               // failure
  }

  MSG msg;
  while (GetMessage(&msg, NULL, 0, 0)) {
    if (IS_STOP_MSG(msg)) break;

    if (msg.message == WM_WTSSESSION_CHANGE) {
      session_change_callback(msg.wParam);
    }

    TranslateMessage(&msg);
    DispatchMessage(&msg);
  }
  ProdCoreStopTracking();
  prod_debug("Tracking stopped\n");
  return 0;
}

int init_tracking() {
  tracking_mutex = CreateMutex(NULL, FALSE, NULL);
  if (tracking_mutex == NULL) return 1;
  return 0;
}

int start_tracking(tracking_opt_t *opts) {
  WaitForSingleObject(tracking_mutex, INFINITE);
  if (ProdCoreIsTracking()) {
    prod_error("tracking started already!\n");
    ReleaseMutex(tracking_mutex);
    return 0;
  }

  tracking_opts = opts;
  if (!(opts->foreground_program || opts->mouse_click || opts->keystroke)) {
    prod_debug("Nothing to be tracked, not doing anything\n");
    ReleaseMutex(tracking_mutex);
    return 1;
  }

  tracking_thread =
      CreateThread(NULL, 0, tracking_loop, NULL, 0, &tracking_thread_id);
  if (tracking_thread == NULL) {
    prod_error("Failed to create tracking thread\n");
    ReleaseMutex(tracking_mutex);
    return 1;
  }

  ReleaseMutex(tracking_mutex);
  return 0;
}

void stop_tracking() {
  WaitForSingleObject(tracking_mutex, INFINITE);
  if (!ProdCoreIsTracking()) {
    prod_error("tracking stopped already!\n");
    return;
  }

  if (!PostThreadMessageA(tracking_thread_id, STOP_MSG, STOP_W_PARAM,
                          STOP_L_PARAM))
    prod_error("Failed to send stop message: %lu\n", GetLastError());

  WaitForSingleObject(tracking_thread, INFINITE);

  tracking_thread = NULL;
  tracking_thread_id = 0;
  tracking_opts = NULL;
  window_handle = NULL;
  mouse_hook = NULL;
  key_hook = NULL;
  window_change_hook = NULL;
  ReleaseMutex(tracking_mutex);
  return;
}

void run_event_loop() {
  event_loop_finished = CreateSemaphore(NULL, 0, 1, NULL);
  if (event_loop_finished == NULL) {
    prod_error("Failed to create sem: %d\n", GetLastError());
  }
  WaitForSingleObject(event_loop_finished, INFINITE);
}

void stop_event_loop() {
  if (!ReleaseSemaphore(event_loop_finished, 1, NULL)) {
    prod_error("ReleaseSemaphore error: %d\n", GetLastError());
  }
}
