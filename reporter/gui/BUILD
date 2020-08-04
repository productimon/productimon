load("@build_bazel_rules_apple//apple:macos.bzl", "macos_application")
load("@com_justbuchanan_rules_qt//:qt.bzl", "qt_cc_library", "qt_resource")
load("@rules_cc//cc:defs.bzl", "cc_binary", "cc_library")
load("@rules_pkg//:pkg.bzl", "pkg_deb", "pkg_tar")

cc_library(
    name = "qt",
    # TODO auto generate a header file containing this
    defines = [
        "LOGO_IMG_PATH=\\\"/images/productimon.png\\\"",
        "SYSTRAY_IMG_PATH=\\\"/images/productimon_white.png\\\"",
    ],
    deps = select({
        "@bazel_tools//src/conditions:darwin": ["@qt_mac_headers//:qt_widgets"],
        "@bazel_tools//src/conditions:darwin_x86_64": ["@qt_mac_headers//:qt_widgets"],
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
    files = [
        "//images:logo",
        "//images:white-logo",
    ],
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
    name = "productimon-reporter",
    deps = [
        ":gui-all",
    ],
)

cc_binary(
    name = "gui-windows-binary",
    srcs = [
        "//images:windows_logo_res",
        "//reporter/plat/windows:version_res",
    ],
    linkopts = select({
        "//internal:debug_build": [],
        "//conditions:default": ["-mwindows"],
    }),
    deps = [":gui-all"],
)

genrule(
    name = "gui-windows",
    srcs = [":gui-windows-binary"],
    outs = ["gui-windows.zip"],
    cmd = "$(location //tools:bundle-win-app.sh) $(location :gui-windows-binary) $(OUTS)",
    message = "Generating Windows bundle",
    tools = ["//tools:bundle-win-app.sh"],
)

macos_application(
    name = "gui-macos",
    additional_contents = {
        "@qt_mac_frameworks//:qt-frameworks": "Frameworks",
        "@qt_mac_plugins//:qt-plugins": "plugins",
    },
    bundle_id = "com.productimon.reporter",
    bundle_name = "Productimon Reporter",
    infoplists = ["//reporter/plat:macos/Info.plist"],
    ipa_post_processor = "//reporter/plat:macos/install_names.sh",
    minimum_os_version = "10.15",  # TODO better than this
    resources = ["//images:apple_icns"],
    deps = [":gui-all"],
)

pkg_tar(
    name = "gui-linux-tar",
    srcs = [":productimon-reporter"],
    mode = "0755",
    package_dir = "/usr/bin/",
)

pkg_deb(
    name = "gui-linux-deb",
    architecture = "amd64",
    data = ":gui-linux-tar",
    depends = [
        "libstdc++6",
        "libdbus-1-3",
        "libx11-6",
        "libxi6",
        "libqt5widgets5",
        "libqt5gui5",
        "libqt5core5a",
        "libc6",
    ],
    description = "Tracks your usage of apps across devices",
    homepage = "https://productimon.com",
    maintainer = "Productimon Team",
    package = "productimon-reporter",
    version = "0.1.0",
)

alias(
    name = "gui",
    actual = select({
        "@bazel_tools//src/conditions:darwin": ":gui-macos",
        "@bazel_tools//src/conditions:windows": ":gui-windows-binary",
        "//conditions:default": ":productimon-reporter",
    }),
    visibility = ["//reporter:__subpackages__"],
)
