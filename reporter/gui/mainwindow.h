#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QtWidgets/QSystemTrayIcon>

#ifndef QT_NO_SYSTEMTRAYICON

#include <QtWidgets/QMainWindow>

#include "reporter/plat/tracking.h"

QT_BEGIN_NAMESPACE
class QCheckBox;
class QGroupBox;
class QPushButton;
class QVBoxLayout;
QT_END_NAMESPACE

class MainWindow : public QMainWindow {
  Q_OBJECT

 public:
  MainWindow();

 private:
  void createActions();
  void createTrayIcon();

  void quit();
  void startStopRecorder();

  void loadSettings();

  void cancelBtnClicked();
  void applyBtnClicked();
  void fetchStates();

  void setupMainLayout();
  void createGridBox();
  void createCB();
  void createButtons();

  QAction *startStopAction;
  QAction *settings;
  QAction *quitAction;

  QPushButton *cancelBtn;
  QPushButton *applyBtn;

  QCheckBox *autoRunOnStartUpCB;
  QCheckBox *foregroungProgCB;
  QCheckBox *mouseClicksCB;
  QCheckBox *keystrokesCB;

  QGroupBox *gridGB;
  QGroupBox *buttonGB;
  QVBoxLayout *mainLayout;

  bool autoRun;
  bool forgroundProg;
  bool mouseClicks;
  bool keystrokes;

  QSystemTrayIcon *trayIcon;
  QMenu *trayIconMenu;

  tracking_opt_t tracking_opts = {
      .foreground_program = 1, .mouse_click = 1, .keystroke = 1};
};

#endif  // QT_NO_SYSTEMTRAYICON

#endif  // MAINWINDOW_H
