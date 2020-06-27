# Productimon DataReporter

## Build Dependencies
### Linux
#### xlib headers
`sudo apt install libx11-dev`


## Build Instructions
### CLI
#### normal build
`bazel build //reporter:reporter_cli`
#### debug build
`bazel build -c dbg //reporter:reporter_cli`

## Running Instructions
`bazel-bin/reporter/reporter_cli HOST:4200 test@productimon.com test`
