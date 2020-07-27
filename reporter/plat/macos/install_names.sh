#!/bin/bash

# TODO do this in a less-painful manner
# define the dependencies and auto gen commands to change the dylib name
# or simply do find and replace

app_name="Productimon Reporter"
base_dir="$1/$app_name.app"
qt_install_dir="/usr/local/opt/qt"
lib_linked_dir="/usr/local/Cellar/qt/5.15.0"

install_name_tool -id "@executable_path/../Frameworks/QtCore.framework/QtCore" \
       "$base_dir/Contents/Frameworks/QtCore.framework/QtCore"
install_name_tool -id "@executable_path/../Frameworks/QtGui.framework/QtGui" \
       "$base_dir/Contents/Frameworks/QtGui.framework/QtGui"
install_name_tool -id "@executable_path/../Frameworks/QtWidgets.framework/QtWidgets"\
       "$base_dir/Contents/Frameworks/QtWidgets.framework/QtWidgets"
install_name_tool -id "@executable_path/../Frameworks/QtDBus.framework/QtDBus"\
       "$base_dir/Contents/Frameworks/QtDBus.framework/QtDBus"
install_name_tool -id "@executable_path/../plugins/platforms/libqcocoa.dylib"\
       "$base_dir/Contents/plugins/platforms/libqcocoa.dylib"
install_name_tool -id "@executable_path/../plugins/styles/libqmacstyle.dylib"\
       "$base_dir/Contents/plugins/styles/libqmacstyle.dylib"

install_name_tool -change "$qt_install_dir/lib/QtCore.framework/Versions/5/QtCore" \
        "@executable_path/../Frameworks/QtCore.framework/QtCore" \
        "$base_dir/Contents/MacOS/$app_name"
install_name_tool -change "$qt_install_dir/lib/QtGui.framework/Versions/5/QtGui" \
        "@executable_path/../Frameworks/QtGui.framework/QtGui" \
        "$base_dir/Contents/MacOS/$app_name"
install_name_tool -change "$qt_install_dir/lib/QtWidgets.framework/Versions/5/QtWidgets" \
        "@executable_path/../Frameworks/QtWidgets.framework/QtWidgets" \
        "$base_dir/Contents/MacOS/$app_name"

install_name_tool -change "$lib_linked_dir/lib/QtCore.framework/Versions/5/QtCore" \
        "@executable_path/../Frameworks/QtCore.framework/QtCore" \
        "$base_dir/Contents/Frameworks/QtGui.framework/QtGui"
install_name_tool -change "$lib_linked_dir/lib/QtCore.framework/Versions/5/QtCore" \
        "@executable_path/../Frameworks/QtCore.framework/QtCore" \
        "$base_dir/Contents/Frameworks/QtWidgets.framework/QtWidgets"
install_name_tool -change "$lib_linked_dir/lib/QtGui.framework/Versions/5/QtGui" \
        "@executable_path/../Frameworks/QtGui.framework/QtGui" \
        "$base_dir/Contents/Frameworks/QtWidgets.framework/QtWidgets"
install_name_tool -change "$lib_linked_dir/lib/QtCore.framework/Versions/5/QtCore" \
        "@executable_path/../Frameworks/QtCore.framework/QtCore" \
        "$base_dir/Contents/Frameworks/QtDBus.framework/QtDBus"
install_name_tool -change "$lib_linked_dir/lib/QtCore.framework/Versions/5/QtCore" \
        "@executable_path/../Frameworks/QtCore.framework/QtCore" \
        "$base_dir/Contents/plugins/platforms/libqcocoa.dylib"
install_name_tool -change "$lib_linked_dir/lib/QtWidgets.framework/Versions/5/QtWidgets" \
        "@executable_path/../Frameworks/QtWidgets.framework/QtWidgets" \
        "$base_dir/Contents/plugins/platforms/libqcocoa.dylib"
install_name_tool -change "$lib_linked_dir/lib/QtGui.framework/Versions/5/QtGui" \
        "@executable_path/../Frameworks/QtGui.framework/QtGui" \
        "$base_dir/Contents/plugins/platforms/libqcocoa.dylib"
install_name_tool -change "$lib_linked_dir/lib/QtPrintSupport.framework/Versions/5/QtPrintSupport" \
        "@executable_path/../Frameworks/QtPrintSupport.framework/QtPrintSupport" \
        "$base_dir/Contents/plugins/platforms/libqcocoa.dylib"
install_name_tool -change "$lib_linked_dir/lib/QtDBus.framework/Versions/5/QtDBus" \
        "@executable_path/../Frameworks/QtDBus.framework/QtDBus" \
        "$base_dir/Contents/plugins/platforms/libqcocoa.dylib"
install_name_tool -change "$lib_linked_dir/lib/QtCore.framework/Versions/5/QtCore" \
        "@executable_path/../Frameworks/QtCore.framework/QtCore" \
        "$base_dir/Contents/Frameworks/QtPrintSupport.framework/QtPrintSupport"
install_name_tool -change "$lib_linked_dir/lib/QtWidgets.framework/Versions/5/QtWidgets" \
        "@executable_path/../Frameworks/QtWidgets.framework/QtWidgets" \
        "$base_dir/Contents/Frameworks/QtPrintSupport.framework/QtPrintSupport"
install_name_tool -change "$lib_linked_dir/lib/QtGui.framework/Versions/5/QtGui" \
        "@executable_path/../Frameworks/QtGui.framework/QtGui" \
        "$base_dir/Contents/Frameworks/QtPrintSupport.framework/QtPrintSupport"
install_name_tool -change "$lib_linked_dir/lib/QtCore.framework/Versions/5/QtCore" \
        "@executable_path/../Frameworks/QtCore.framework/QtCore" \
        "$base_dir/Contents/plugins/styles/libqmacstyle.dylib"
install_name_tool -change "$lib_linked_dir/lib/QtWidgets.framework/Versions/5/QtWidgets" \
        "@executable_path/../Frameworks/QtWidgets.framework/QtWidgets" \
        "$base_dir/Contents/plugins/styles/libqmacstyle.dylib"
install_name_tool -change "$lib_linked_dir/lib/QtGui.framework/Versions/5/QtGui" \
        "@executable_path/../Frameworks/QtGui.framework/QtGui" \
        "$base_dir/Contents/plugins/styles/libqmacstyle.dylib"
