load("@rules_cc//cc:defs.bzl", "cc_library")

config_setting(
    name = "debug_build",
    values = {
        "compilation_mode": "dbg",
    },
)

cc_library(
    name = "trackinglib-linux",
    srcs = glob([
        "linux/*.c",
        "linux/*.h",
    ]),
    hdrs = ["tracking.h"],
    defines = select({
        "debug_build": ["DEBUG"],
        "//conditions:default": [],
    }),
    visibility = ["//reporter:__pkg__"],
    deps = [
        "//reporter/core:core.cc",
        "@dbus",
        "@system_libs//:pthread",
        "@system_libs//:x11",
        "@system_libs//:xi",
    ],
)

cc_library(
    name = "trackinglib-windows",
    srcs = glob([
        "windows/*.c",
        "windows/*.h",
    ]),
    hdrs = ["tracking.h"],
    defines = select({
        "debug_build": ["DEBUG"],
        "//conditions:default": [],
    }),
    linkopts = [
        "-luser32",
        "-lversion",
        "-lshlwapi",
        "-lwtsapi32",
    ],
    visibility = ["//reporter:__pkg__"],
    deps = [
        "//reporter/core:core.cc",
    ],
)
