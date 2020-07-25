# browser extension

Generic browser extension code is here.

Browser-specific hooks to get data are under `plat/chrome`

## Build

We currently only support building a Chrome extension.

```
bazel build -c dbg //reporter/browser:chrome_extension
```

This will generate a zip file that you can upload to Chrome extension store.

To install it locally, unzip it, go to `chrome://extensions` and use `Pack extension` to select the unzipped directory.
