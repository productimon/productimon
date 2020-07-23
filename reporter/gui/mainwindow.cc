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

#include "reporter/core/cgo/cgo.h"
#include "reporter/plat/tracking.h"

MainWindow::MainWindow() {
  autoRun = true;
  forgroundProg = true;
  mouseClicks = true;
  keystrokes = true;

  setupMainLayout();

  setWindowTitle(tr("Settings"));

  createActions();
  createTrayIcon();

  QWidget *widget = new QWidget();
  widget->setLayout(mainLayout);
  setCentralWidget(widget);

  trayIcon->show();
  trayIcon->showMessage("Productimon Data Reporter", "Authenticated");

  resize(400, 300);
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
    trayIcon->showMessage("Productimon Data Reporter", "Tracking stopped");
    startStopAction->setText(tr("Start Recorder"));
  } else {
    start_tracking(&tracking_opts);
    trayIcon->showMessage("Productimon Data Reporter", "Tracking started");
    startStopAction->setText(tr("Stop Recorder"));
  }
}

void MainWindow::createActions() {
  startStopAction = new QAction(tr("&Start Recorder"), this);
  connect(startStopAction, &QAction::triggered, this,
          &MainWindow::startStopRecorder);

  settings = new QAction(tr("&Settings"), this);
  connect(settings, &QAction::triggered, this, &MainWindow::loadSettings);

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

  trayIcon->setIcon(QIcon(":/reporter/gui/images/nucleusIcon.png"));
}

void MainWindow::createCB() {
  autoRunOnStartUpCB = new QCheckBox(this);
  foregroungProgCB = new QCheckBox(this);
  mouseClicksCB = new QCheckBox(this);
  keystrokesCB = new QCheckBox(this);
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

  gridLayout->addWidget(new QLabel(tr("Foreground Programs:")), 0, 0);
  gridLayout->addWidget(new QLabel(tr("Mouse Clicks:")), 1, 0);
  gridLayout->addWidget(new QLabel(tr("Keystrokes:")), 2, 0);
  gridLayout->addWidget(new QLabel(tr("Auto Run at Startup:")), 3, 0);

  gridLayout->addWidget(foregroungProgCB, 0, 2);
  gridLayout->addWidget(mouseClicksCB, 1, 2);
  gridLayout->addWidget(keystrokesCB, 2, 2);
  gridLayout->addWidget(autoRunOnStartUpCB, 3, 2);

  gridLayout->setColumnMinimumWidth(0, 300);

  for (int i = 0; i < 4; i++) gridLayout->setRowMinimumHeight(i, 20);

  gridGB->setLayout(gridLayout);
}

void MainWindow::loadSettings() {
  fetchStates();
  this->show();
  QWidget::activateWindow();
  raise();
}

void MainWindow::fetchStates() {
  // find current recorder settings

  autoRunOnStartUpCB->setChecked(autoRun);
  foregroungProgCB->setChecked(forgroundProg);
  mouseClicksCB->setChecked(mouseClicks);
  keystrokesCB->setChecked(keystrokes);
}

void MainWindow::cancelBtnClicked() { this->hide(); }

void MainWindow::applyBtnClicked() {
  // submit changes

  autoRun = autoRunOnStartUpCB->checkState();
  forgroundProg = foregroungProgCB->checkState();
  mouseClicks = mouseClicksCB->checkState();
  keystrokes = keystrokesCB->checkState();

  this->hide();
}

#endif
