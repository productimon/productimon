#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QtWidgets/QSystemTrayIcon>

#ifndef QT_NO_SYSTEMTRAYICON

#include <QtWidgets/QMainWindow>
#include <vector>

#include "reporter/gui/OptionCheckBox.h"
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
  void startStopReporter();

  void showSettingWindow();

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

  std::vector<OptionCheckBox *> checkBoxes;

  QGroupBox *gridGB;
  QGroupBox *buttonGB;
  QVBoxLayout *mainLayout;

  QSystemTrayIcon *trayIcon;
  QMenu *trayIconMenu;
};

#endif  // QT_NO_SYSTEMTRAYICON

#endif  // MAINWINDOW_H
