#include <QtWidgets/QApplication>

#include "reporter/core/cgo/cgo.h"
#include "reporter/gui/loginwindow.h"
#include "reporter/gui/mainwindow.h"
#include "reporter/plat/tracking.h"

int main(int argc, char *argv[]) {
  ProdCoreReadConfig();

  QApplication app(argc, argv);

  app.setQuitOnLastWindowClosed(false);

  if (init_tracking()) {
    prod_error("Failed to init tracking\n");
    return 1;
  }

  /* try login using cert key first */
  if (ProdCoreInitReporterByCert()) {
    MainWindow mainwindow;
    return app.exec();
  }

  /* fall back to login window */
  LoginWindow loginWindow;
  loginWindow.show();
  return app.exec();
}
