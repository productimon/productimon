load("@io_bazel_rules_go//extras:embed_data.bzl", "go_embed_data")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_embed_data(
    name = "db_data",
    srcs = [
        "schema.sql",
    ],
    flatten = True,
    package = "db",
    visibility = ["//visibility:public"],
)

# keep
go_library(
    name = "go_default_library",
    srcs = [":db_data"],
    importpath = "git.yiad.am/productimon/aggregator/db",
    visibility = ["//visibility:public"],
)
