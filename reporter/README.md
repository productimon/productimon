# Productimon DataReporter

## Build Dependencies

### Linux

`sudo apt install gcc g++ python libx11-dev libxi-dev cmake libexpat1-dev qt5-default imagemagick`

### macOS

```
xcode-select --install
brew install qt imagemagick
```

After installing qt, make sure to update your PATH to include qt's bin. For me it was `echo 'export PATH="/usr/local/opt/qt/bin:$PATH"' >> ~/.bashrc`

### Windows

Follow docs/windows.md first.

Then run `pacman -S mingw-w64-x86_64-qt-creator mingw-w64-x86_64-imagemagick`

## Build Instructions

### normal build

- CLI: `bazel build //reporter/cli`
- GUI: `bazel build //reporter/gui`

### debug build

- CLI: `bazel -c dbg build //reporter/cli`
- GUI: `bazel -c dbg build //reporter/gui`

## Runtime Dependencies

### Linux

`sudo apt install libstdc++6 libc6 libx11-6 libxi6 libdbus-1-3 libqt5widgets5 libqt5gui5 libqt5core5a`

### macOS

It should work on lastest macOS without any runtime dependencies

## Running Instructions

Usually, you don't need any CLI argument. Just run `bazel run //reporter/cli` or `bazel run //reporter/gui`

See `bazel run //reporter/cli -- --help` or `bazel run //reporter/gui -- --help` for more info
