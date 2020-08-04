exports_files(
    glob(["*"]),
    visibility = ["//visibility:public"],
)

alias(
    name = "logo",
    actual = ":productimon.png",
    visibility = ["//visibility:public"],
)

alias(
    name = "white-logo",
    actual = ":productimon_white.png",
    visibility = ["//visibility:public"],
)

genrule(
    name = "apple_icns",
    srcs = [
        ":logo",
    ],
    outs = ["icons.icns"],
    cmd = "./$(location //tools:gen-apple-icns.sh) $(location :logo) $(OUTS)",
    message = "Generating Apple icons",
    tools = ["//tools:gen-apple-icns.sh"],
    visibility = ["//reporter/gui:__subpackages__"],
)

genrule(
    name = "webfe_ico",
    srcs = [":logo"],
    outs = ["favicon.ico"],
    cmd = select({
        "@bazel_tools//src/conditions:windows": "magick convert $(location :logo) -define icon:auto-resize=\"128,64,32,16\" $(OUTS)",
        "//conditions:default": "convert $(location :logo) -define icon:auto-resize=\"128,64,32,16\" $(OUTS)",
    }),
    message = "Generating favicon.ico",
    visibility = ["//viewer/webfe:__subpackages__"],
)

genrule(
    name = "windows_logo_res",
    srcs = [
        ":logo",
        "//reporter/plat/windows:logo.rc",
    ],
    outs = ["logo.o"],
    cmd = "./$(location //tools:gen-windows-ico.sh) $(location //reporter/plat/windows:logo.rc) $(location :logo) $(OUTS)",
    message = "Generating Windows logo resource file",
    tools = ["//tools:gen-windows-ico.sh"],
    visibility = ["//reporter/gui:__subpackages__"],
)
