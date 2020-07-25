document.addEventListener("DOMContentLoaded", function () {
  var coreModule = chrome.extension.getBackgroundPage();
  toggleEnabledUI(coreModule.userTracking);

  document.querySelector("#about").addEventListener("click", function () {
    window.open("https://github.com/productimon/productimon");
  });

  document.querySelector("#start").addEventListener("click", function () {
    coreModule.userStartTracking();
    toggleEnabledUI(true);
  });

  document.querySelector("#stop").addEventListener("click", function () {
    coreModule.userStopTracking();
    toggleEnabledUI(false);
  });

  function toggleEnabledUI(enabled) {
    document.querySelector("#start").classList.toggle("hide", enabled);
    document.querySelector("#stop").classList.toggle("hide", !enabled);
  }
});
