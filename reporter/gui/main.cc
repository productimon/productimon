#include <QtWidgets/QApplication>

#include "reporter/core/core.h"
#include "reporter/gui/loginwindow.h"
#include "reporter/gui/mainwindow.h"
#include "reporter/plat/tracking.h"

int main(int argc, char *argv[]) {
  QApplication app(argc, argv);

  app.setQuitOnLastWindowClosed(false);

  if (init_tracking()) {
    prod_error("Failed to init tracking\n");
    return 1;
  }

  ReadConfig();

  /* try login using cert key first */
  if (InitReporterByCert()) {
    MainWindow mainwindow;
    return app.exec();
  }

  /* fall back to login window */
  LoginWindow loginWindow;
  loginWindow.show();
  return app.exec();
}
