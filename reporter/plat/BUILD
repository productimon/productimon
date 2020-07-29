load("@rules_cc//cc:defs.bzl", "cc_library", "objc_library")

cc_library(
    name = "trackinglib-generic",
    srcs = ["tracking.c"],
    hdrs = ["tracking.h"],
    deps = ["//reporter/core/cgo:cgo.cc"],
)

cc_library(
    name = "trackinglib-linux",
    srcs = glob([
        "linux/*.c",
        "linux/*.h",
    ]),
    defines = select({
        "//internal:debug_build": ["DEBUG"],
        "//conditions:default": [],
    }),
    deps = [
        ":trackinglib-generic",
        "@dbus",
        "@system_libs//:pthread",
        "@system_libs//:x11",
        "@system_libs//:xi",
    ],
)

cc_library(
    name = "trackinglib-windows",
    srcs = ["//reporter/plat/windows:source_files"],
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
    deps = [
        ":trackinglib-generic",
    ],
)

objc_library(
    name = "trackinglib-macos",
    srcs = ["macos/tracking.m"],
    defines = select({
        "//internal:debug_build": ["DEBUG"],
        "//conditions:default": [],
    }),
    sdk_frameworks = [
        "AppKit",
        "Security",
    ],
    deps = [
        ":trackinglib-generic",
    ],
)

cc_library(
    name = "trackinglib",
    visibility = ["//reporter:__subpackages__"],
    deps = select({
        "@bazel_tools//src/conditions:darwin": [":trackinglib-macos"],
        "@bazel_tools//src/conditions:darwin_x86_64": [":trackinglib-macos"],
        "@bazel_tools//src/conditions:linux_x86_64": [":trackinglib-linux"],
        "@bazel_tools//src/conditions:windows": [":trackinglib-windows"],
    }),
)

exports_files(
    [
        "macos/Info.plist",
        "macos/install_names.sh",
    ],
)
