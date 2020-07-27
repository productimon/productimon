load("@build_bazel_rules_apple//apple:macos.bzl", "macos_application")
load("@com_justbuchanan_rules_qt//:qt.bzl", "qt_cc_library", "qt_resource")
load("@rules_cc//cc:defs.bzl", "cc_binary", "cc_library")

cc_library(
    name = "qt",
    deps = select({
        "@bazel_tools//src/conditions:darwin": ["@qt_mac//:qt_widgets"],
        "@bazel_tools//src/conditions:darwin_x86_64": ["@qt_mac//:qt_widgets"],
        "@bazel_tools//src/conditions:linux_x86_64": ["@qt_linux//:qt_widgets"],
        "@bazel_tools//src/conditions:windows": ["@qt_msys//:qt_widgets"],
    }),
)

qt_cc_library(
    name = "loginwindow",
    srcs = ["loginwindow.cc"],
    hdrs = ["loginwindow.h"],
    deps = [
        ":mainwindow",
        ":qt",
        "//reporter/plat:trackinglib",
    ],
)

qt_cc_library(
    name = "mainwindow",
    srcs = ["mainwindow.cc"],
    hdrs = [
        "OptionCheckBox.h",
        "mainwindow.h",
    ],
    deps = [
        ":images",
        ":qt",
        "//reporter/plat:trackinglib",
    ],
)

qt_resource(
    name = "images",
    files = ["//images:nucleusIcon.png"],
)

cc_library(
    name = "gui-all",
    srcs = ["main.cc"],
    deps = [
        ":loginwindow",
        ":mainwindow",
        ":qt",
        "//reporter/plat:trackinglib",
    ],
)

cc_binary(
    name = "gui-binary",
    deps = [
        ":gui-all",
    ],
)

macos_application(
    name = "gui-macos",
    bundle_id = "com.productimon.reporter",
    bundle_name = "Productimon Reporter",
    infoplists = ["//reporter/plat:macos/Info.plist"],
    minimum_os_version = "10.15",  # TODO better than this
    resources = ["//images:apple_icns"],
    deps = [":gui-all"],
)

alias(
    name = "gui",
    actual = select({
        "@bazel_tools//src/conditions:darwin": ":gui-macos",
        "//conditions:default": ":gui-binary",
    }),
    visibility = ["//reporter:__subpackages__"],
)
