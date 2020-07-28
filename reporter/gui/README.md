# Productimon DataReporter GUI

## Build Dependencies

### Linux

#### Qt
TODO

### macOS

#### xcode build tools
xcode-select --install

#### Qt
`brew install qt`

After installing qt, make sure to update your PATH to include qt's bin. For me it was `echo 'export PATH="/usr/local/opt/qt/bin:$PATH"' >> ~/.bashrc`

This dependency is also needed for running our GUI since we're using a dynamic linked version of Qt now.

### Windows
#### Qt in MSYS
`pacman -S mingw-w64-x86_64-qt-creator`

See details in [https://wiki.qt.io/MSYS2](https://wiki.qt.io/MSYS2)

`bazel build //reporter/gui` will create a standalone binary that requires dlls from the msys environment and Qt

`bazel build //reporter/gui:gui-windows` will create a zip file containing the binary and all the necessary supporting files required to run on a Windows machine at the location `bazel-bin/reporter/gui/gui-windows.zip`
