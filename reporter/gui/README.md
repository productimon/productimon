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

This dependency is also needed for running our GUI since we're using a dynamic linked version of Qt now.
