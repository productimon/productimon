load("@bazel_gazelle//:def.bzl", "gazelle")
load("@com_github_bazelbuild_buildtools//buildifier:def.bzl", "buildifier")

# gazelle:prefix git.yiad.am/productimon
gazelle(
    name = "gazelle",
)

buildifier(
    name = "buildifier",
    exclude_patterns = [
        "./bazel-*/*",
        "/node_modules/*",
    ],
    lint_mode = "fix",
    lint_warnings = ["all"],
)
