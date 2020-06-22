config_setting(
    name = "debug_build",
    values = {
        "compilation_mode": "dbg",
    },
)

cc_library(
    name = "trackinglib-linux",
    srcs = glob(['linux/*.c']),
    hdrs = ['tracking.h'],
    deps = ["@system_libs//:x11", "@system_libs//:pthread"],
    defines = select({
        "debug_build": ['DEBUG'],
        "//conditions:default": [],
    }),
    visibility = ["//reporter:__pkg__"],
)
