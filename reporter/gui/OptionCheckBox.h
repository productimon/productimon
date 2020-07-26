#pragma once

#include <QtWidgets/QCheckBox>
#include <QtWidgets/QWidget>

class OptionCheckBox : public QCheckBox {
  Q_OBJECT
 public:
  const char *optName;
  const char *displayName;

  OptionCheckBox(QWidget *parent, const char *_optName,
                 const char *_displayName)
      : QCheckBox{parent}, optName{_optName}, displayName{_displayName} {}

  OptionCheckBox(QWidget *parent = nullptr)
      : OptionCheckBox{parent, "null", "null"} {}
};
