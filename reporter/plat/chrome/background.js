// when user switches to a non-browser window
// isTracking() returns false but userTracking is still true
var userTracking = false;

function userStartTracking() {
  startTracking((success) => {
    userTracking = success;
  });
}

function userStopTracking() {
  stopTracking(() => {
    userTracking = false;
  });
}

function onCoreLoaded() {
  console.log("js: got callback from core loaded");
  // TODO: remove this
  login("api.productimon.com:4201", "test@productimon.com", "test", "chrome");
}

function initCoreModule() {
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
    console.log("WebAssembly is not supported in your browser");
  }
}

function initHooks() {
  chrome.tabs.onUpdated.addListener(function (tabId, changeInfo, tab) {
    console.log("updated " + tab.url);
    if (userTracking) switchUrl(tab.url);
  });

  chrome.tabs.onActivated.addListener(function (activeInfo) {
    chrome.tabs.get(activeInfo.tabId, function (tab) {
      console.log("on highlight " + tab.url + " " + window.location.href);
      if (userTracking) switchUrl(tab.url);
    });
  });

  chrome.windows.onFocusChanged.addListener(function (windowId) {
    if (windowId >= 0) {
      if (userTracking) {
        // it's safe to call startTracking when tracking already started - it doesn't do anything
        // this is to ensure that if we switched back from non-browser to browser, we'll resume tracking
        startTracking((success) => {
          if (!success) {
            console.log("Failed to start tracking");
            return;
          }
          chrome.tabs.query(
            {
              active: true,
              windowId: windowId,
            },
            function (tabarr) {
              if (tabarr.length == 1) {
                console.log("window focus change observed: " + tabarr[0].url);
                switchUrl(tabarr[0].url);
              }
            }
          );
        });
      }
    } else {
      // we're no longer on a browser window
      stopTracking();
    }
  });
}

function init() {
  initCoreModule();
  initHooks();
}

init();
