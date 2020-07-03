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

## Build Instructions

### CLI

#### normal build

`bazel build //reporter:reporter_cli`

#### debug build

`bazel build -c dbg //reporter:reporter_cli`

## Runtime Dependencies

### Linux

#### xlib client

`sudo apt install libx11-6`

#### x input extension

`sudo apt install libxi6`

## Running Instructions

### CLI

`bazel-bin/reporter/reporter_cli HOST:4200 test@productimon.com test`
