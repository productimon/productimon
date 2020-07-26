#include "reporter/gui/loginwindow.h"

#include <QtWidgets/QtWidgets>
#include <string>
#include <cstdlib>

#include "reporter/core/cgo/cgo.h"
#include "reporter/gui/mainwindow.h"

#define TO_C_STR(qtstr) (qtstr.toUtf8().data())

LoginWindow::LoginWindow() {
  createGridGroupBox();
  createBtns();
  createBtnGrid();
  createAndSetMainLayout();

  setWindowTitle(tr("Data Reporter Login"));

  resize(400, 300);
}

void LoginWindow::tryLogin() {
  QString server = serverField->text();
  QString username = usernameField->text();
  QString password = passwordField->text();
  QString deviceName = deviceNameField->text();

  prod_debug("login using: %s %s %s %s\n", TO_C_STR(server), TO_C_STR(username),
             TO_C_STR(password), TO_C_STR(deviceName));
  if (ProdCoreInitReporterByCreds(TO_C_STR(server), TO_C_STR(username),
                                  TO_C_STR(password), TO_C_STR(deviceName))) {
    auto settingsWindow = new MainWindow();
    settingsWindow->show();
    this->hide();
  } else {
    // TODO figure out a way to pass error message from core to here
    QMessageBox::warning(this, "Login failed", "Failed to login");
  }
}

void LoginWindow::reject() { QApplication::quit(); }

void LoginWindow::createAndSetMainLayout() {
  QVBoxLayout *mainLayout = new QVBoxLayout;
  mainLayout->addWidget(gridGroupBox);
  mainLayout->addWidget(btnGrid);
  setLayout(mainLayout);
}

void LoginWindow::createBtnGrid() {
  btnGrid = new QGroupBox();
  QGridLayout *layout = new QGridLayout;

  layout->addWidget(loginBtn, 1, 0);
  layout->addWidget(quitBtn, 1, 1);
  btnGrid->setLayout(layout);
}

void LoginWindow::createBtns() {
  loginBtn = new QPushButton(tr("&Login"));
  quitBtn = new QPushButton(tr("&Quit"));
  connect(quitBtn, &QAbstractButton::clicked, this, &LoginWindow::quit);
  connect(loginBtn, &QAbstractButton::clicked, this, &LoginWindow::tryLogin);
}

void LoginWindow::quit() { QApplication::quit(); }

void LoginWindow::createGridGroupBox() {
  gridGroupBox = new QGroupBox();
  QGridLayout *layout = new QGridLayout;
  char *server = ProdCoreGetServer();
  serverField = new QLineEdit(server);
  free(server);
  usernameField = new QLineEdit;
  passwordField = new QLineEdit;
  passwordField->setEchoMode(QLineEdit::Password);
  deviceNameField = new QLineEdit;
  labels[0] = new QLabel(tr("Server: "));
  labels[1] = new QLabel(tr("Username: "));
  labels[2] = new QLabel(tr("Password: "));
  labels[3] = new QLabel(tr("Device Name: "));
  layout->addWidget(labels[0], 1, 0);
  layout->addWidget(labels[1], 2, 0);
  layout->addWidget(labels[2], 3, 0);
  layout->addWidget(labels[3], 4, 0);

  layout->addWidget(serverField, 1, 1);
  layout->addWidget(usernameField, 2, 1);
  layout->addWidget(passwordField, 3, 1);
  layout->addWidget(deviceNameField, 4, 1);

  layout->setColumnStretch(1, 10);
  gridGroupBox->setLayout(layout);
}
