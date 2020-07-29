genrule(
    name = "version_res",
    srcs = ["version.rc"],
    outs = ["version.o"],
    cmd = "windres $(location version.rc) -O coff -o $(OUTS)",
    message = "Generating Windows version resource file",
    visibility = ["//reporter/gui:__subpackages__"],
)

filegroup(
    name = "source_files",
    srcs = glob([
        "*.c",
        "*.h",
    ]),
    visibility = ["//reporter/plat:__subpackages__"],
)

exports_files(glob(["*.rc"]))
