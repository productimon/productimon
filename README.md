<img src="images/productimon.png" width="150">

# Productimon

This monorepo contains all code and documentation for Productimon, a cross-platform activity/usage/screen time tracking and analytics tool. It currently supports Linux, Windows, macOS, and Chrome OS/Chromium OS, with a cross-platform core codebase that's easy to extend to other platforms.

<img src="images/screenshot.png" width="100%">


## Architecture

Productimon is composed of 4 modules:

- [**DataAggregator**](aggregator), key component to connect everything together, serves as RPC server and store all data;
- [**DataAnalyzer**](analyzer), analytical pipeline to predict app label and transform raw event stream to interval index (bundled in DataAggregator);
- [**DataReporter**](reporter), cross-platform native app as well as browser extension to report activities to designated server;
- [**DataViewer**](viewer), front-end to view activity analytics and metrics (bundled in DataAggregator).

## Installing

We are releasing a set of compiled binaries on GitHub for your convenience.

On Windows: unzip our binary release [productimon-reporter-windows-x86_64.zip](https://github.com/productimon/productimon/releases/download/v0.1/productimon-reporter-windows-x86_64.zip) and run reporter.exe within the unzipped folder

On Linux: install the deb package [productimon-reporter-linux-x86_64.deb](https://github.com/productimon/productimon/releases/download/v0.1/productimon-reporter-linux-x86_64.deb) and run productimon-reporter in your terminal. Alternatively, download the executable [productimon-reporter-linux-x86_64](https://github.com/productimon/productimon/releases/download/v0.1/productimon-reporter-linux-x86_64) and run it directly. NOTE you might need to install runtime dependencies before running the program, see details in [reporter/README.md](reporter/README.md)

On macOS: unzip our binary release [productimon-reporter-darwin-x86_64.zip](https://github.com/productimon/productimon/releases/download/v0.1/productimon-reporter-darwin-x86_64.zip) and install the unzipped app to /Applications after which you can find it and run it in LaunchPad.

On Google Chrome: unzip the release [productimon-reporter-chrome_extension.zip](https://github.com/productimon/productimon/releases/download/v0.1/productimon-reporter-chrome_extension.zip) and navigate to chrome://extensions/, enable developer mode in the top right corner and click load unpacked button to select the unzipped extension folder to install it.


We are hosting a public DataAggregator instance at [my.productimon.com](https://my.productimon.com), but
you can also deploy your own server to fully customize it and keep your data in your own hands.

## Building

Check the README.md file under each directory to check more details and required dev/runtime dependencies.

DataAggregator, DataAnalyzer, and DataViewer are bundled into a single binary:

```
bazel build //aggregator
```

For DataReporter,
```
bazel build //reporter/cli
bazel build //reporter/gui
bazel build //reporter/gui:gui-windows
bazel build //reporter/gui:gui-linux-deb
bazel build //reporter/browser:chrome_extension
```

Insert `-c dbg` after `bazel build` to get a debug build.

## Development

### Source Code

We use our internal Gerrit instance for code review as the canonical repo. We also host a
[public GitHub mirror for open-sourcing](https://github.com/productimon/productimon)


### Coding Style

- Use `buildifier` for Bazel BUILD files.
- Use `clang-format` for c/c++/proto files.
- Use `gofmt` for Golang files.
- Use `prettier` for HTML/CSS/JS/Markdown files.

### Design Docs

- [System Architecture Diagram](https://docs.google.com/drawings/d/1TWAxg70vu_xM8AbtSZTbwLAWlQPJRRmYKtxeFkw_HkI)
- [API Documentation](/proto)


### Dev Env Setup

- [Windows Setup](docs/windows.md)

## Contributing

You are welcome to contribute! Just send in a Pull Request on GitHub and we'll sync it to our internal Gerrit
for code review and merging.

All submissions, including submissions by project members, require review.

## LICENSE

Open-sourced with 💗 under [Apache 2.0 License](LICENSE).

```
Copyright 2020 The Productimon Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

## Disclaimer

This is still alpha-quality software. All APIs are subject to breaking changes. Use at your own risk.

There's no SLA guarantee for the public instance at [my.productimon.com](https://my.productimon.com). If you use this public instance, you agree to our [Privacy Policy](PRIVACY_POLICY.md). You don't have to agree to it if you deploy your own server.
