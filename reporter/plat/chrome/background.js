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
    switchUrl(tab.url);
  });

  chrome.tabs.onActivated.addListener(function (activeInfo) {
    chrome.tabs.get(activeInfo.tabId, function (tab) {
      console.log("on highlight " + tab.url + " " + window.location.href);
      switchUrl(tab.url);
    });
  });

  chrome.windows.onFocusChanged.addListener(function (windowId) {
    if (windowId >= 0) {
      // we're on a browser window
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
    } else {
      // we're no longer on a browser window
      // TODO: pause tracking
    }
  });
}

function init() {
  initCoreModule();
  initHooks();
}

init();
