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

```
bazel-bin/aggregator/aggregator_/aggregator --help
```

When running for the first time, server certificate is automatically generated. Database is also automatically initiated.
The credentials for the initial admin user will be printed out in console.

Visit `http://127.0.0.1:4201/`
