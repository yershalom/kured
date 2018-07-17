package(default_visibility = ["//visibility:public"])

load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["azure.go"],
    importmap = "github.com/weaveworks/kured/vendor/k8s.io/client-go/plugin/pkg/client/auth/azure",
    importpath = "k8s.io/client-go/plugin/pkg/client/auth/azure",
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/net:go_default_library",
        "//vendor/k8s.io/client-go/rest:go_default_library",
        "@com_github_azure_go_autorest//autorest:go_default_library",
        "@com_github_azure_go_autorest//autorest/adal:go_default_library",
        "@com_github_azure_go_autorest//autorest/azure:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
)
