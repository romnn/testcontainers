load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@bazel_gazelle//:def.bzl", "gazelle")

# gazelle:prefix github.com/romnnn/testcontainers
gazelle(name = "gazelle")

go_library(
    name = "go_default_library",
    srcs = [
        "log.go",
        "merge.go",
        "network.go",
        "options.go",
        "testcontainers.go",
        "utils.go",
    ],
    importpath = "github.com/romnnn/testcontainers",
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_cenkalti_backoff_v4//:go_default_library",
        "@com_github_google_uuid//:go_default_library",
        "@com_github_imdario_mergo//:go_default_library",
        "@com_github_prometheus_common//log:go_default_library",
        "@com_github_romnnn_testcontainers_go//:go_default_library",
    ],
)