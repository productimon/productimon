load("@com_justbuchanan_rules_qt//:qt.bzl", "qt_cc_library", "qt_resource")
load("@rules_cc//cc:defs.bzl", "cc_binary", "cc_library")

cc_library(
    name = "auto-qt",
    deps = select({
        "@bazel_tools//src/conditions:darwin": ["@qt_mac//:qt_widgets"],
        "@bazel_tools//src/conditions:linux_x86_64": ["@qt_linux//:qt_widgets"],
        "@bazel_tools//src/conditions:windows": ["@qt_msys//:qt_widgets"],
    }),
)

qt_cc_library(
    name = "loginwindow",
    srcs = ["loginwindow.cc"],
    hdrs = ["loginwindow.h"],
    deps = [
        ":auto-qt",
        ":mainwindow",
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
        ":auto-qt",
        ":images",
        "//reporter/plat:trackinglib",
    ],
)

qt_resource(
    name = "images",
    files = ["//images:nucleusIcon.png"],
)

cc_binary(
    name = "gui",
    srcs = ["main.cc"],
    deps = [
        ":auto-qt",
        ":loginwindow",
        ":mainwindow",
        "//reporter/plat:trackinglib",
    ],
)
