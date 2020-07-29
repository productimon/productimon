# Productimon DataReporter

## Build Dependencies

### Linux

#### xlib headers

`sudo apt install libx11-dev`

#### x input headers

`sudo apt install libxi-dev`

#### cmake

`sudo apt install cmake`

#### libexpat headers

`sudo apt install libexpat1-dev`

#### Qt

TODO

### macOS

#### xcode build tools

xcode-select --install

#### Qt

`brew install qt`

After installing qt, make sure to update your PATH to include qt's bin. For me it was `echo 'export PATH="/usr/local/opt/qt/bin:$PATH"' >> ~/.bashrc`

### Windows

#### Qt in MSYS

See docs/windows.md for more details.

`pacman -S mingw-w64-x86_64-qt-creator`

See details in [https://wiki.qt.io/MSYS2](https://wiki.qt.io/MSYS2)

This dependency is also needed for running our GUI since we're using a dynamic linked version of Qt now.

#### imagemagick
To generate ico icon files

`pacman -S mingw-w64-x86_64-imagemagick`

## Build Instructions

### normal build

- CLI: `bazel build //reporter/cli`
- GUI: `bazel build //reporter/gui`

### debug build

- CLI: `bazel -c dbg build //reporter/cli`
- GUI: `bazel -c dbg build //reporter/gui`

## Runtime Dependencies

### Linux

#### xlib client

`sudo apt install libx11-6`

#### x input extension

`sudo apt install libxi6`

### macOS

It should work on lastest macOS without any runtime dependencies

## Running Instructions

### CLI

Usually, you don't need any CLI argument. Just run `bazel run //reporter/cli` or `bazel run //reporter/gui`

See `bazel run //reporter/cli -- --help` or `bazel run //reporter/gui -- --help` for more info
