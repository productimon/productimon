load("@rules_cc//cc:defs.bzl", "cc_library")

cc_library(
    name = "qt_core",
    hdrs = glob(["QtCore/**"]),
    includes = ["."],
    linkopts = [
        "-F/usr/local/opt/qt/lib",
        "-framework QtCore",
    ],
)

cc_library(
    name = "qt_gui",
    hdrs = glob(["QtGui/**"]),
    includes = ["."],
    linkopts = [
        "-F/usr/local/opt/qt/lib",
        "-framework QtGui",
    ],
    deps = [":qt_core"],
)

cc_library(
    name = "qt_widgets",
    hdrs = glob(["QtWidgets/**"]),
    includes = ["."],
    linkopts = [
        "-F/usr/local/opt/qt/lib",
        "-framework QtWidgets",
    ],
    visibility = ["//visibility:public"],
    deps = [
        ":qt_core",
        ":qt_gui",
    ],
)
