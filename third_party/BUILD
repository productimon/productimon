load("@io_bazel_rules_closure//closure:defs.bzl", "closure_js_library")

licenses(["notice"])

exports_files(
    glob(["*"]),
    visibility = ["//visibility:public"],
)

closure_js_library(
    name = "wasm_exec",
    srcs = [
        "wasm_exec.js",
    ],
    suppress = [
        "JSC_UNKNOWN_EXPR_TYPE",
        "JSC_NAME_REFERENCE_IN_EXTERNS",
        "checkTypes",
    ],
    visibility = ["//visibility:public"],
)

closure_js_library(
    name = "chrome_externs",
    srcs = [
        "@chrome_externs//:chrome.js",
        "@chrome_externs//:chrome_extensions.js",
    ],
    visibility = ["//visibility:public"],
)
