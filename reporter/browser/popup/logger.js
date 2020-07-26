goog.module("productimon.reporter.browser.popup.logger");

const Console = goog.require("goog.debug.Console");
const LogManager = goog.require("goog.debug.LogManager");
const Logger = goog.require("goog.debug.Logger");

/**
 * Init logging
 */
function initialize() {
  const console = new Console();
  console.setCapturing(true);
  LogManager.getRoot().setLevel(Logger.Level.FINEST);
}

exports = { initialize };
