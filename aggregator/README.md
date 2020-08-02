# aggregator

### compile

This bundles the frontend as well.

**prod build** (minified js)

```
bazel build //aggregator
```

**dev build** (unminified js with sourcemap, ~10MB)

```
bazel build -c dbg //aggregator
```

### run

Most of the configuration is provided as command line arguments. Check help before running.

```
bazel-bin/aggregator/aggregator_/aggregator --help
```

When running for the first time, server certificate is automatically generated. Database is also automatically initiated.
The credentials for the initial admin user will be printed out in console.

Aggregator listens on three different ports: HTTP, HTTPS, and gRPC (with mTLS encryption).

The HTTP port is used for redirecting to HTTPS. None of your personal data is transmitted over HTTP or without encryption.

You can provide your own HTTPS certificate. If not provided, one will be automatically provisioned with Let's Encrypt.
See help for more details. HTTP is also used for ACME protocol to automatically provision your certificate from Let's Encrypt.

Aggregator provides three ways to access its APIs:

- gRPC (with mTLS certificate)
- gRPC (with mTLS certificate) over HTTPS WebSocket
- gRPC-Web with auth token over HTTPS
