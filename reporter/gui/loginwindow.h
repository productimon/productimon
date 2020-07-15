#ifndef LOGINWINDOW_H
#define LOGINWINDOW_H

#include <QtWidgets/QDialog>

QT_BEGIN_NAMESPACE
class QGroupBox;
class QLabel;
class QLineEdit;
class QMenuBar;
QT_END_NAMESPACE

class LoginWindow : public QDialog {
  Q_OBJECT

 public:
  LoginWindow();

 private:
  void createGridGroupBox();
  void quit();
  void tryLogin();
  void createBtns();
  void createBtnGrid();
  void createAndSetMainLayout();
  void reject() override;

  enum { NumGridRows = 4 };

  QGroupBox *gridGroupBox;
  QGroupBox *btnGrid;
  QLabel *labels[NumGridRows];

  QPushButton *loginBtn;
  QPushButton *quitBtn;

  QLineEdit *serverField;
  QLineEdit *usernameField;
  QLineEdit *passwordField;
  QLineEdit *deviceNameField;
};

#endif  // LoginWindow_H
