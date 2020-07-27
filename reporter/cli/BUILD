load("@rules_cc//cc:defs.bzl", "cc_binary")

cc_binary(
    name = "cli",
    srcs = ["cli.c"],
    visibility = ["//reporter:__subpackages__"],
    deps = ["//reporter/plat:trackinglib"],
)
