goog.module("productimon.reporter.browser.popup");

const login = goog.require("productimon.reporter.browser.popup.login");
const logger = goog.require("productimon.reporter.browser.popup.logger");

logger.initialize();
login.init();
