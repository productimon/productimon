load("@rules_cc//cc:defs.bzl", "cc_binary")

cc_binary(
    name = "reporter_cli",
    srcs = ["cli.c"],
    deps = select({
        "@bazel_tools//src/conditions:darwin": ["//reporter/plat:trackinglib-macos"],
        "@bazel_tools//src/conditions:linux_x86_64": ["//reporter/plat:trackinglib-linux"],
        "@bazel_tools//src/conditions:windows": ["//reporter/plat:trackinglib-windows"],
    }) + ["//reporter/plat:tracking-header"],
)
