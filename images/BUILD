exports_files(
    glob(["*"]),
    visibility = ["//visibility:public"],
)

alias(
    name = "logo",
    actual = ":nucleusIcon.png",
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

# TODO generate apple icns file automatically
