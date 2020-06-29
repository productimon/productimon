load("@rules_foreign_cc//tools/build_defs:cmake.bzl", "cmake_external")

filegroup(
    name = "all",
    srcs = glob(["**"]),
)

cmake_external(
    name = "dbus",
    lib_source = ":all",
    out_include_dirs = [
        "include/dbus-1.0",
        "lib/dbus-1.0/include",
    ],
    shared_libraries = ["libdbus-1.so"],
    # TODO figure out a way to build static lib and link
    # it to our linux trackinglib
    visibility = ["//visibility:public"],
)
