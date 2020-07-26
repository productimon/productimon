goog.module("productimon.reporter.plat.chrome");

const Console = goog.require("goog.debug.Console");
const LogManager = goog.require("goog.debug.LogManager");
const Logger = goog.require("goog.debug.Logger");
const log = goog.require("goog.log");
const logger = log.getLogger("productimon.reporter.plat.chrome");

class Reporter {
  constructor() {
    // when user switches to a non-browser window
    // isTracking() returns false but userTracking is still true
    /** @type {boolean} */
    this.userTracking = false;

    /** @type {boolean} */
    this.loggedIn = false;

    /** @type {string} */
    this.os = "";
    chrome.runtime.getPlatformInfo(
      function (/** chrome.runtime.PlatformInfo */ info) {
        this.os = info.os;
      }.bind(this)
    );
  }

  /**
   * Called when user clicks Start tracking
   */
  userStartTracking() {
    log.info(logger, "user clicked start tracking");
    startTracking((/** boolean */ success) => {
      log.info(
        logger,
        "start tracking callback from core, success: " + success
      );
      this.userTracking = success;
    });
  }

  /**
   * Called when user clicks Stop tracking
   */
  userStopTracking() {
    log.info(logger, "user clicked stop tracking");
    stopTracking(() => {
      log.info(logger, "stop tracking callback from core");
      this.userTracking = false;
    });
  }

  /**
   * Start up core module wasm
   */
  initCoreModule() {
    if (WebAssembly) {
      // WebAssembly.instantiateStreaming is not currently available in Safari
      if (WebAssembly && !WebAssembly.instantiateStreaming) {
        // polyfill
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
          const source = await (await resp).arrayBuffer();
          return await WebAssembly.instantiate(source, importObject);
        };
      }

      const go = new Go();
      WebAssembly.instantiateStreaming(
        fetch("productimon.wasm"),
        go.importObject
      ).then((result) => {
        go.run(result.instance);
      });
    } else {
      log.error(logger, "WebAssembly is not supported in your browser");
    }
  }

  /**
   * https://developer.chrome.com/extensions/tabs#event-onUpdated
   * @param {number} tabId
   * @param {Object} changeInfo
   * @param {Tab} tab
   */
  onUpdated(tabId, changeInfo, tab) {
    if (this.userTracking) {
      log.info(logger, "updated " + tab.url);
      switchUrl(tab.url);
    }
  }

  /**
   * https://developer.chrome.com/extensions/tabs#event-onActivated
   * @param {{tabId: number, windowId: number}} activeInfo
   */
  onActivated(activeInfo) {
    if (this.userTracking) {
      chrome.tabs.get(activeInfo.tabId, function (tab) {
        log.info(logger, "on highlight " + tab.url);
        switchUrl(tab.url);
      });
    }
  }

  /**
   * https://developer.chrome.com/extensions/windows#event-onFocusChanged
   * @param {number} windowId
   */
  onFocusChanged(windowId) {
    if (this.userTracking) {
      log.info(logger, "onFocusChanged: " + windowId);
      // chrome.windows.onFocusChanged triggers -1 on Linux/CrOS/Windows
      // under many weird circumstances when we still have focus
      // see crbug/387377#c30
      if (windowId >= 0 || this.os == "mac") {
        this.switchToWindow(windowId);
      }
    }
  }

  /**
   * Called when we switch to a new window.
   * This assumes we're currently tracking
   *
   * @param {number} windowId the new window ID, negative to pause tracking
   */
  switchToWindow(windowId) {
    if (windowId >= 0) {
      // it's safe to call startTracking when tracking already started - it doesn't do anything
      // this is to ensure that if we switched back from non-browser to browser, we'll resume tracking
      startTracking((/** boolean */ success) => {
        if (!success) {
          log.error(logger, "Failed to start tracking");
          return;
        }
        chrome.tabs.query(
          {
            active: true,
            windowId: windowId,
          },
          function (tabarr) {
            if (tabarr.length == 1) {
              log.info(logger, "switch to: " + tabarr[0].url);
              switchUrl(tabarr[0].url);
            } else {
              // TODO: check if this is actually possible and if this is the correct way to deal with this
              log.error(
                logger,
                "i don't know why there're multiple active tabs in single window - pause tracking"
              );
              stopTracking();
            }
          }
        );
      });
    } else {
      // we're no longer on a browser window
      log.info(logger, "lose focus - stop tracking");
      stopTracking();
    }
  }

  /**
   * https://developer.chrome.com/extensions/runtime#event-onMessage
   * @param {*} message
   * @param {MessageSender} sender
   * @param {function(*)} sendResponse
   * @return {boolean} whether to send response async
   */
  onMessage(message, sender, sendResponse) {
    const request = /** @type {{action: string, payload: *}} */ (message);
    switch (request.action) {
      case "STOP_TRACKING":
        this.userStopTracking();
        return false;
      case "START_TRACKING":
        this.userStartTracking();
        return false;
      case "GET_TRACKING":
        sendResponse(this.userTracking);
        return false;
      case "CHECK_LOGIN":
        sendResponse(this.loggedIn);
        return false;
      case "LOGIN":
        const payload = /** @type {{serverName: string, username: string, password: string, deviceName: string}} */ (request.payload);
        login(
          payload["serverName"],
          payload["username"],
          payload["password"],
          payload["deviceName"],
          (/** boolean */ success) => {
            this.loggedIn = success;
            sendResponse(success);
          }
        );
        return true;
    }
    return false;
  }

  /**
   * Install hooks used to track user
   */
  initHooks() {
    chrome.tabs.onUpdated.addListener(this.onUpdated.bind(this));
    chrome.tabs.onActivated.addListener(this.onActivated.bind(this));
    chrome.runtime.onMessage.addListener(this.onMessage.bind(this));
    // chrome.windows.onFocusChanged is unreliable on Linux/CrOS/Windows
    // so we fall back to polling chrome.windows.getLastFocused
    // see crbug/387377#c30
    chrome.windows.onFocusChanged.addListener(this.onFocusChanged.bind(this));
    if (this.os != "mac") this.startPollingWindow();
  }

  /**
   * Poll focused window to generate events
   */
  startPollingWindow() {
    window.setInterval(
      function () {
        // https://developer.chrome.com/extensions/windows#method-getLastFocused
        chrome.windows.getLastFocused(
          null,
          function (/** ChromeWindow */ w) {
            if (this.userTracking) {
              this.switchToWindow(w.focused ? w.id : -1);
            }
          }.bind(this)
        );
      }.bind(this),
      1000
    );
  }

  /**
   * init everything
   */
  init() {
    this.initCoreModule();
    this.initHooks();
  }
}

/**
 * main entry point
 */
function init() {
  const console = new Console();
  console.setCapturing(true);
  LogManager.getRoot().setLevel(Logger.Level.FINEST);

  const reporter = new Reporter();
  window["onCoreLoaded"] = function (/** boolean */ loggedIn) {
    log.info(logger, "js: got callback from core loaded " + loggedIn);
    reporter.loggedIn = loggedIn;
  };
  reporter.init();
}

goog.exportSymbol("productimon.reporter.plat.chrome.init", init);

exports = { Reporter };
