<img src="images/productimon.png" width="150">

# Productimon

This monorepo contains all code and documentation for Productimon, a cross-platform activity/usage/screen time tracking
and analytics tool.


## Installing

TBA

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
