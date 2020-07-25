load("@rules_cc//cc:defs.bzl", "cc_library", "objc_library")

cc_library(
    name = "tracking-header",
    hdrs = ["tracking.h"],
    visibility = ["//reporter:__subpackages__"],
)

cc_library(
    name = "trackinglib-linux",
    srcs = glob([
        "linux/*.c",
        "linux/*.h",
    ]),
    hdrs = ["tracking.h"],
    defines = select({
        "//internal:debug_build": ["DEBUG"],
        "//conditions:default": [],
    }),
    visibility = ["//reporter:__pkg__"],
    deps = [
        "//reporter/core/cgo:cgo.cc",
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
        "//internal:debug_build": ["DEBUG"],
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
        "//reporter/core/cgo:cgo.cc",
    ],
)

objc_library(
    name = "trackinglib-macos",
    srcs = ["macos/tracking.m"],
    hdrs = ["tracking.h"],
    defines = select({
        "//internal:debug_build": ["DEBUG"],
        "//conditions:default": [],
    }),
    sdk_frameworks = [
        "AppKit",
        "Security",
    ],
    visibility = ["//reporter:__pkg__"],
    deps = [
        "//reporter/core/cgo:cgo.cc",
    ],
)

cc_library(
    name = "trackinglib",
    visibility = ["//reporter:__subpackages__"],
    deps = select({
        "@bazel_tools//src/conditions:darwin": ["//reporter/plat:trackinglib-macos"],
        "@bazel_tools//src/conditions:linux_x86_64": ["//reporter/plat:trackinglib-linux"],
        "@bazel_tools//src/conditions:windows": ["//reporter/plat:trackinglib-windows"],
    }),
)
