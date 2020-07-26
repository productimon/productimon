load("@io_bazel_rules_closure//closure:defs.bzl", "filegroup_external")

def setup_chrome_externs():
    filegroup_external(
        name = "chrome_externs",
        licenses = ["notice"],  # Apache 2.0
        sha256_urls = {
            "ab31bb7ec7a8c094a4103753e63235a9d672d99eb3790eb1a66768d138626382": [
                "https://raw.githubusercontent.com/google/closure-compiler/791d5b71fa5562556839e3c4d3f3c63e738ad101/contrib/externs/chrome.js",
            ],
            "b2831aade6b45d248a218a914e92317e07751c099adea13a67c29bb4c41a846c": [
                "https://raw.githubusercontent.com/google/closure-compiler/791d5b71fa5562556839e3c4d3f3c63e738ad101/contrib/externs/chrome_extensions.js",
            ],
        },
    )
