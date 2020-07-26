goog.module("productimon.reporter.browser.popup.options");

const dom = goog.require("goog.dom");
const events = goog.require("goog.events");
const log = goog.require("goog.log");
const templates = goog.require(
  "productimon.reporter.browser.popup.options.templates"
);
const logger = log.getLogger("productimon.reporter.browser.popup.options");

/** @type {!Array<{id: number, name: string, title: string, explaination: string, init: function()}>} */
const options = [
  {
    id: 0,
    name: "on",
    title: "On",
    explaination: "Report websites you use to server.",
    init: () => {
      toggleEnabledUI(true);
    },
  },
  {
    id: 1,
    name: "off",
    title: "Off",
    explaination: "Nothing gets sent to server.",
    init: () => {
      toggleEnabledUI(false);
    },
  },
];

/**
 * entry point
 */
function init() {
  log.info(logger, "options init");
  dom.getElement("main").innerHTML = templates.options({
    options: options,
  });
  chrome.extension.sendMessage({ action: "GET_TRACKING" }, toggleEnabledUI);
  const optiondoms = dom.getElementsByClass("option");
  for (var i = 0; i < optiondoms.length; i++) {
    events.listen(optiondoms[i], events.EventType.CLICK, options[i].init);
  }
}

/**
 * @param {boolean} enabled
 */
function toggleEnabledUI(enabled) {
  log.info(logger, "toggleEnabledUI: " + enabled);
  const title = dom.getElement("title");
  const header = dom.getElement("header");
  if (enabled) {
    chrome.extension.sendMessage({ action: "START_TRACKING" });
    header.style.backgroundColor = "#1EBEA5";
    dom.setTextContent(title, "Websites are being sent to Productimon");
    dom.getElement("on").checked = true;
  } else {
    chrome.extension.sendMessage({ action: "STOP_TRACKING" });
    header.style.backgroundColor = "#404040";
    dom.setTextContent(title, "Productimon is not tracking");
    dom.getElement("off").checked = true;
  }
}

exports = { init };
