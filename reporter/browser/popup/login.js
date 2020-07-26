goog.module("productimon.reporter.browser.popup.login");

const dom = goog.require("goog.dom");
const events = goog.require("goog.events");
const forms = goog.require("goog.dom.forms");
const log = goog.require("goog.log");
const options = goog.require("productimon.reporter.browser.popup.options");
const templates = goog.require(
  "productimon.reporter.browser.popup.login.templates"
);
const logger = log.getLogger("productimon.reporter.browser.popup.login");

/**
 * entry point
 */
function init() {
  log.info(logger, "login init");
  chrome.extension.sendMessage({ action: "CHECK_LOGIN" }, (
    /** boolean */ loggedIn
  ) => {
    log.info(logger, "is logged in: " + loggedIn);
    if (loggedIn) {
      options.init();
    } else {
      dom.getElement("main").innerHTML = templates.login();
      dom.getElement("header").style.backgroundColor = "#D5572B";
      dom.setTextContent(dom.getElement("title"), "Login to Productimon");
      events.listen(dom.getElement("login"), events.EventType.CLICK, () => {
        var payload = {};
        // otherwise, renaming will break integration with background.js
        payload["serverName"] = forms.getValue(dom.getElement("serverName"));
        payload["username"] = forms.getValue(dom.getElement("username"));
        payload["password"] = forms.getValue(dom.getElement("password"));
        payload["deviceName"] = forms.getValue(dom.getElement("deviceName"));
        chrome.extension.sendMessage(
          {
            action: "LOGIN",
            payload: payload,
          },
          (/** boolean */ success) => {
            log.info(logger, "login success: " + success);
            if (success) {
              options.init();
            }
          }
        );
      });
    }
  });
}

exports = { init };
