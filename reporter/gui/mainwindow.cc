#include "reporter/gui/mainwindow.h"

#ifndef QT_NO_SYSTEMTRAYICON

#include <QtWidgets/QAction>
#include <QtWidgets/QApplication>
#include <QtWidgets/QCheckBox>
#include <QtWidgets/QGroupBox>
#include <QtWidgets/QLabel>
#include <QtWidgets/QMainWindow>
#include <QtWidgets/QMenu>
#include <QtWidgets/QPushButton>
#include <QtWidgets/QVBoxLayout>
#include <vector>

#include "reporter/core/cgo/cgo.h"
#include "reporter/gui/OptionCheckBox.h"
#include "reporter/plat/tracking.h"

// TODO rename file names to corespond to class name
MainWindow::MainWindow() {
  setupMainLayout();

  fetchStates();

  setWindowTitle(tr("Settings"));

  createActions();
  createTrayIcon();

  QWidget *widget = new QWidget();
  widget->setLayout(mainLayout);
  setCentralWidget(widget);

  trayIcon->show();
  // TODO set the icon to be our logo
  trayIcon->showMessage("Productimon Data Reporter", "Authenticated");

  this->setFixedSize(QSize(400, 300));
}

void MainWindow::setupMainLayout() {
  mainLayout = new QVBoxLayout();
  createCB();
  createButtons();
  createGridBox();

  mainLayout->addWidget(gridGB);
  mainLayout->addSpacing(10);
  mainLayout->addWidget(buttonGB);
}

void MainWindow::quit() {
  stop_tracking();
  ProdCoreQuitReporter();
  QApplication::quit();
}

void MainWindow::startStopRecorder() {
  if (ProdCoreIsTracking()) {
    stop_tracking();
    // TODO set the icon to be our logo
    trayIcon->showMessage("Productimon Data Reporter", "Tracking stopped");
    startStopAction->setText(tr("Start Recorder"));
  } else {
    if (start_tracking()) {
      // TODO set the icon to be our logo
      // TODO use a message box instead?
      // figure out a way to pass error message from core to here
      // something like ProdCoreGetLastError would work fine
      trayIcon->showMessage("Productimon Data Reporter",
                            "Failed to start tracking");
      return;
    }
    // TODO set the icon to be our logo
    trayIcon->showMessage("Productimon Data Reporter", "Tracking started");
    startStopAction->setText(tr("Stop Recorder"));
  }
}

void MainWindow::createActions() {
  startStopAction = new QAction(tr("&Start Recorder"), this);
  connect(startStopAction, &QAction::triggered, this,
          &MainWindow::startStopRecorder);

  settings = new QAction(tr("&Settings"), this);
  connect(settings, &QAction::triggered, this, &MainWindow::showSettingWindow);

  quitAction = new QAction(tr("&Quit"), this);
  connect(quitAction, &QAction::triggered, this, &MainWindow::quit);
}

void MainWindow::createTrayIcon() {
  if (!QSystemTrayIcon::isSystemTrayAvailable())
    prod_debug(
        "Warning: QSystemTrayIcon::isSystemTrayAvailable returned 0. System "
        "tray icon might not show up properly in your system.");
  trayIconMenu = new QMenu(this);

  trayIconMenu->addAction(startStopAction);
  trayIconMenu->addAction(settings);
  trayIconMenu->addAction(quitAction);

  trayIcon = new QSystemTrayIcon(this);
  trayIcon->setContextMenu(trayIconMenu);

  trayIcon->setIcon(QIcon(":/images/nucleusIcon.png"));
}

void MainWindow::createCB() {
  for (size_t i = 0; i < NUM_OPTIONS; i++) {
    auto optName = tracking_options[i].opt_name;
    auto displayName = tracking_options[i].display_name;
    checkBoxes.push_back(new OptionCheckBox(this, optName, displayName));
  }
}

void MainWindow::createButtons() {
  cancelBtn = new QPushButton("Cancel");
  applyBtn = new QPushButton("Apply");

  connect(cancelBtn, &QAbstractButton::clicked, this,
          &MainWindow::cancelBtnClicked);
  connect(applyBtn, &QAbstractButton::clicked, this,
          &MainWindow::applyBtnClicked);

  QHBoxLayout *buttonLayout = new QHBoxLayout;

  buttonLayout->addWidget(cancelBtn);
  buttonLayout->addWidget(applyBtn);

  buttonGB = new QGroupBox();

  buttonGB->setLayout(buttonLayout);
}

void MainWindow::createGridBox() {
  gridGB = new QGroupBox(tr("Tracking Options"));
  QGridLayout *gridLayout = new QGridLayout();

  for (size_t i = 0; i < checkBoxes.size(); i++) {
    auto checkbox = checkBoxes[i];
    gridLayout->addWidget(new QLabel(tr(checkbox->displayName)), i, 0);
    gridLayout->addWidget(checkbox, i, 2);
  }

  gridLayout->setColumnMinimumWidth(0, 300);

  for (int i = 0; i < 4; i++) gridLayout->setRowMinimumHeight(i, 20);

  gridGB->setLayout(gridLayout);
}

void MainWindow::showSettingWindow() {
  fetchStates();
  this->show();
  QWidget::activateWindow();
  raise();
}

void MainWindow::fetchStates() {
  for (auto checkbox : checkBoxes) {
    checkbox->setChecked(get_option(checkbox->optName));
  }
}

void MainWindow::cancelBtnClicked() { this->hide(); }

void MainWindow::applyBtnClicked() {
  std::vector<const char *> options;
  bool tracking = ProdCoreIsTracking();

  if (tracking) {
    stop_tracking();
  }

  for (auto checkbox : checkBoxes) {
    if (checkbox->checkState()) options.push_back(checkbox->optName);
  }
  options.push_back(NULL);

  ProdCoreSetOptions((char **)options.data());
  ProdCoreSaveConfig();

  if (tracking) {
    start_tracking();
  }
  this->hide();
  // TODO set the icon to be our logo
  trayIcon->showMessage("Productimon Data Reporter", "Settings updated");
}

#endif
