workspace(
    name = "productimon",
    managed_directories = {"@npm": ["node_modules"]},
)

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

# Protobuf library

http_archive(
    name = "com_google_protobuf",
    sha256 = "e5265d552e12c1f39c72842fa91d84941726026fa056d914ea6a25cd58d7bbf8",
    strip_prefix = "protobuf-3.12.3",
    urls = ["https://github.com/google/protobuf/archive/v3.12.3.zip"],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

# Go & Gazelle rules

# NOTE(adamyi): we need 585a27ad0ab5bdd185aa3bd5b0877a778d4777ad and dfd0b678d384b75d57a5e1a326099a54273de849 which is not yet in a stable relase
http_archive(
    name = "io_bazel_rules_go",
    sha256 = "c1d2e55123334f377495002401ad6fe0415ec96fd322a741b7ec12a05f2aaafc",
    strip_prefix = "rules_go-dfd0b678d384b75d57a5e1a326099a54273de849",
    urls = [
        "https://github.com/bazelbuild/rules_go/archive/dfd0b678d384b75d57a5e1a326099a54273de849.tar.gz",
    ],
)

# local_repository(
#     name = "io_bazel_rules_go",
#     path = "../rules_go",
# )

http_archive(
    name = "bazel_gazelle",
    sha256 = "cdb02a887a7187ea4d5a27452311a75ed8637379a1287d8eeb952138ea485f7d",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

http_archive(
    name = "com_github_bazelbuild_buildtools",
    sha256 = "a0e79f5876a1552ae8000882e4189941688f359a80b2bc1d7e3a51cab6257ba1",
    strip_prefix = "buildtools-3.0.0",
    url = "https://github.com/bazelbuild/buildtools/archive/3.0.0.tar.gz",
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

# proto

http_archive(
    name = "rules_proto",
    sha256 = "602e7161d9195e50246177e7c55b2f39950a9cf7366f74ed5f22fd45750cd208",
    strip_prefix = "rules_proto-97d8af4dc474595af3900dd85cb3a29ad28cc313",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_proto/archive/97d8af4dc474595af3900dd85cb3a29ad28cc313.tar.gz",
        "https://github.com/bazelbuild/rules_proto/archive/97d8af4dc474595af3900dd85cb3a29ad28cc313.tar.gz",
    ],
)

load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies", "rules_proto_toolchains")

rules_proto_dependencies()

rules_proto_toolchains()

NODEJS_COMMIT = "acc81d7c03d57ef8d7daa99f51e9b0d5c76b34ac"

NODEJS_SHA256 = "ce868fd3490042d7fff058e2a050a5dc4fb7896ecc5aa113fe4c2c3e57ddd05d"

# nodejs
http_archive(
    name = "build_bazel_rules_nodejs",
    sha256 = NODEJS_SHA256,
    strip_prefix = "rules_nodejs-" + NODEJS_COMMIT,
    urls = ["https://github.com/adamyi/rules_nodejs/archive/" + NODEJS_COMMIT + ".tar.gz"],
)

#local_repository(
#    name = "build_bazel_rules_nodejs",
#    path = "../rules_nodejs",
#)

load("@build_bazel_rules_nodejs//:package.bzl", "rules_nodejs_dev_dependencies")

rules_nodejs_dev_dependencies()

#local_repository(
#    name = "npm_bazel_typescript",
#    path = "../rules_nodejs/packages/typescript/src",
#)

http_archive(
    name = "npm_bazel_typescript",
    sha256 = NODEJS_SHA256,
    strip_prefix = "rules_nodejs-" + NODEJS_COMMIT + "/packages/typescript/src",
    urls = ["https://github.com/adamyi/rules_nodejs/archive/" + NODEJS_COMMIT + ".tar.gz"],
)

load("@build_bazel_rules_nodejs//:index.bzl", "yarn_install")

yarn_install(
    name = "npm",
    package_json = "//:package.json",
    yarn_lock = "//:yarn.lock",
)

http_archive(
    name = "build_bazel_rules_typescript",
    sha256 = "5ac7855faddf296c33f4b1f8f4917d0800bda6cb3f8c133ba097bbe00e149d90",
    strip_prefix = "rules_typescript-5d79e42953eb8614d961ccf0e3440884e974eeb3",
    url = "https://github.com/bazelbuild/rules_typescript/archive/5d79e42953eb8614d961ccf0e3440884e974eeb3.tar.gz",
)

load("@build_bazel_rules_typescript//:package.bzl", "rules_typescript_dev_dependencies")

rules_typescript_dev_dependencies()

load("@npm//:install_bazel_dependencies.bzl", "install_bazel_dependencies")

install_bazel_dependencies()

load("@build_bazel_rules_typescript//internal:ts_repositories.bzl", "ts_setup_dev_workspace")

ts_setup_dev_workspace()

load("@npm_bazel_typescript//:index.bzl", "ts_setup_workspace")

ts_setup_workspace()

go_repository(
    name = "org_golang_google_grpc",
    importpath = "google.golang.org/grpc",
    sum = "h1:EC2SB8S04d2r73uptxphDSUG+kTKVgjRPF+N3xpxRB4=",
    version = "v1.29.1",
)

go_repository(
    name = "org_golang_x_net",
    importpath = "golang.org/x/net",
    sum = "h1:0GoQqolDA55aaLxZyTzK/Y2ePZzZTUrRacwib7cNsYQ=",
    version = "v0.0.0-20190404232315-eb5bcb51f2a3",
)

go_repository(
    name = "org_golang_x_text",
    importpath = "golang.org/x/text",
    sum = "h1:tW2bmiBqwgJj/UpqtC8EpXEZVYOwU0yG4iWbprSVAcs=",
    version = "v0.3.2",
)

# reporter linux libs
new_local_repository(
    name = "system_libs",
    build_file = "third_party/system_libs.BUILD",
    path = "/usr/lib/x86_64-linux-gnu",
)

http_archive(
    name = "rules_foreign_cc",
    sha256 = "7d96526be03eff25dde14c51fa6b1088f7f94f9c643082132ab0759addb8b363",
    strip_prefix = "rules_foreign_cc-2f8c9b999f68d225af195763d4c7f0e7bd63cd70",
    urls = ["https://github.com/Chester-P/rules_foreign_cc/archive/2f8c9b999f68d225af195763d4c7f0e7bd63cd70.tar.gz"],
)

load("@rules_foreign_cc//:workspace_definitions.bzl", "rules_foreign_cc_dependencies")

rules_foreign_cc_dependencies()

http_archive(
    name = "dbus",
    build_file = "@//third_party:dbus.BUILD",
    sha256 = "f56b0aa015d0cd13e235225484f411e3c587a0f852c12da03852a324dd1cafb3",
    strip_prefix = "dbus-1.13.16",
    urls = ["https://dbus.freedesktop.org/releases/dbus/dbus-1.13.16.tar.xz"],
)

go_repository(
    name = "com_github_dgrijalva_jwt_go",
    importpath = "github.com/dgrijalva/jwt-go",
    sum = "h1:7qlOGliEKZXTDg6OTjfoBKDXWrumCAMpl/TFQ4/5kLM=",
    version = "v3.2.0+incompatible",
)

go_repository(
    name = "com_github_mattn_go_sqlite3",
    importpath = "github.com/mattn/go-sqlite3",
    sum = "h1:jbhqpg7tQe4SupckyijYiy0mJJ/pRyHvXf7JdWK860o=",
    version = "v1.10.0",
)

go_repository(
    name = "org_golang_x_crypto",
    importpath = "golang.org/x/crypto",
    sum = "h1:HuIa8hRrWRSrqYzx1qI49NNxhdi2PrY7gxVSq1JjLDc=",
    version = "v0.0.0-20190701094942-4def268fd1a4",
)

go_repository(
    name = "com_github_improbable_eng_grpc_web",
    importpath = "github.com/improbable-eng/grpc-web",
    sum = "h1:GlCS+lMZzIkfouf7CNqY+qqpowdKuJLSLLcKVfM1oLc=",
    version = "v0.12.0",
)

go_repository(
    name = "com_github_gorilla_websocket",
    importpath = "github.com/gorilla/websocket",
    sum = "h1:+/TMaTYc4QFitKJxsQ7Yye35DkWvkdLcvGKqM+x0Ufc=",
    version = "v1.4.2",
)

go_repository(
    name = "com_github_rs_cors",
    importpath = "github.com/rs/cors",
    sum = "h1:+88SsELBHx5r+hZ8TCkggzSstaWNbDvThkVK8H6f9ik=",
    version = "v1.7.0",
)

go_repository(
    name = "com_github_desertbit_timer",
    importpath = "github.com/desertbit/timer",
    sum = "h1:U5y3Y5UE0w7amNe7Z5G/twsBW0KEalRQXZzf8ufSh9I=",
    version = "v0.0.0-20180107155436-c41aec40b27f",
)

go_repository(
    name = "com_github_google_uuid",
    importpath = "github.com/google/uuid",
    sum = "h1:Gkbcsh/GbpXz7lPftLA3P6TYMwjCLYm83jiFQZF/3gY=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_cloudflare_cfssl",
    importpath = "github.com/cloudflare/cfssl",
    sum = "h1:vScfU2DrIUI9VPHBVeeAQ0q5A+9yshO1Gz+3QoUQiKw=",
    version = "v1.4.1",
)

go_repository(
    name = "com_github_akavel_rsrc",
    importpath = "github.com/akavel/rsrc",
    sum = "h1:zjWn7ukO9Kc5Q62DOJCcxGpXC18RawVtYAGdz2aLlfw=",
    version = "v0.8.0",
)

go_repository(
    name = "com_github_certifi_gocertifi",
    importpath = "github.com/certifi/gocertifi",
    sum = "h1:6/yVvBsKeAw05IUj4AzvrxaCnDjN4nUqKjW9+w5wixg=",
    version = "v0.0.0-20180118203423-deb3ae2ef261",
)

go_repository(
    name = "com_github_cloudflare_backoff",
    importpath = "github.com/cloudflare/backoff",
    sum = "h1:8d1CEOF1xldesKds5tRG3tExBsMOgWYownMHNCsev54=",
    version = "v0.0.0-20161212185259-647f3cdfc87a",
)

go_repository(
    name = "com_github_cloudflare_go_metrics",
    importpath = "github.com/cloudflare/go-metrics",
    sum = "h1:/8sZyuGTAU2+fYv0Sz9lBcipqX0b7i4eUl8pSStk/4g=",
    version = "v0.0.0-20151117154305-6a9aea36fb41",
)

go_repository(
    name = "com_github_cloudflare_redoctober",
    importpath = "github.com/cloudflare/redoctober",
    sum = "h1:p0Q1GvgWtVf46XpMMibupKiE7aQxPYUIb+/jLTTK2kM=",
    version = "v0.0.0-20171127175943-746a508df14c",
)

go_repository(
    name = "com_github_daaku_go_zipexe",
    importpath = "github.com/daaku/go.zipexe",
    sum = "h1:VSOgZtH418pH9L16hC/JrgSNJbbAL26pj7lmD1+CGdY=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_davecgh_go_spew",
    importpath = "github.com/davecgh/go-spew",
    sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_geertjohan_go_incremental",
    importpath = "github.com/GeertJohan/go.incremental",
    sum = "h1:7AH+pY1XUgQE4Y1HcXYaMqAI0m9yrFqo/jt0CW30vsg=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_geertjohan_go_rice",
    importpath = "github.com/GeertJohan/go.rice",
    sum = "h1:KkI6O9uMaQU3VEKaj01ulavtF7o1fWT7+pk/4voiMLQ=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_getsentry_raven_go",
    importpath = "github.com/getsentry/raven-go",
    sum = "h1:ELaJ1cjF2nEJeIlHXahGme22yG7TK+3jB6IGCq0Cdrc=",
    version = "v0.0.0-20180121060056-563b81fc02b7",
)

go_repository(
    name = "com_github_go_sql_driver_mysql",
    importpath = "github.com/go-sql-driver/mysql",
    sum = "h1:pgwjLi/dvffoP9aabwkT3AKpXQM93QARkjFhDDqC1UE=",
    version = "v1.3.0",
)

go_repository(
    name = "com_github_golang_protobuf",
    importpath = "github.com/golang/protobuf",
    sum = "h1:YF8+flBXS5eO826T4nzqPrxfhQThhXl0YzfuUPu4SBg=",
    version = "v1.3.1",
)

go_repository(
    name = "com_github_google_certificate_transparency_go",
    importpath = "github.com/google/certificate-transparency-go",
    sum = "h1:Yf1aXowfZ2nuboBsg7iYGLmwsOARdV86pfH3g95wXmE=",
    version = "v1.0.21",
)

go_repository(
    name = "com_github_jessevdk_go_flags",
    importpath = "github.com/jessevdk/go-flags",
    sum = "h1:4IU2WS7AumrZ/40jfhf4QVDMsQwqA7VEHozFRrGARJA=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_jmhodges_clock",
    importpath = "github.com/jmhodges/clock",
    sum = "h1:dYTbLf4m0a5u0KLmPfB6mgxbcV7588bOCx79hxa5Sr4=",
    version = "v0.0.0-20160418191101-880ee4c33548",
)

go_repository(
    name = "com_github_jmoiron_sqlx",
    importpath = "github.com/jmoiron/sqlx",
    sum = "h1:ryslCsfLTV4Cm/9NXqCJirlbYodWqFiTH454IaSn/fY=",
    version = "v0.0.0-20180124204410-05cef0741ade",
)

go_repository(
    name = "com_github_kisielk_sqlstruct",
    importpath = "github.com/kisielk/sqlstruct",
    sum = "h1:o/c0aWEP/m6n61xlYW2QP4t9424qlJOsxugn5Zds2Rg=",
    version = "v0.0.0-20150923205031-648daed35d49",
)

go_repository(
    name = "com_github_kisom_goutils",
    importpath = "github.com/kisom/goutils",
    sum = "h1:z4HEOgAnFq+e1+O4QdVsyDPatJDu5Ei/7w7DRbYjsIA=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_konsorten_go_windows_terminal_sequences",
    importpath = "github.com/konsorten/go-windows-terminal-sequences",
    sum = "h1:mweAR1A6xJ3oS2pRaGiHgQ4OO8tzTaLawm8vnODuwDk=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_kr_pretty",
    importpath = "github.com/kr/pretty",
    sum = "h1:L/CwN0zerZDmRFUapSPitk6f+Q3+0za1rQkzVuMiMFI=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_kr_pty",
    importpath = "github.com/kr/pty",
    sum = "h1:VkoXIwSboBpnk99O/KFauAEILuNHv5DVFKZMBN/gUgw=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_kr_text",
    importpath = "github.com/kr/text",
    sum = "h1:45sCR5RtlFHMR4UwH9sdQ5TC8v0qDQCHnXt+kaKSTVE=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_kylelemons_go_gypsy",
    importpath = "github.com/kylelemons/go-gypsy",
    sum = "h1:mkl3tvPHIuPaWsLtmHTybJeoVEW7cbePK73Ir8VtruA=",
    version = "v0.0.0-20160905020020-08cad365cd28",
)

go_repository(
    name = "com_github_lib_pq",
    importpath = "github.com/lib/pq",
    sum = "h1:Ou506ViB5uo2GloKFWIYi5hwRJn4AAOXuLVv8RMY9+4=",
    version = "v0.0.0-20180201184707-88edab080323",
)

go_repository(
    name = "com_github_mreiferson_go_httpclient",
    importpath = "github.com/mreiferson/go-httpclient",
    sum = "h1:oKIteTqeSpenyTrOVj5zkiyCaflLa8B+CD0324otT+o=",
    version = "v0.0.0-20160630210159-31f0106b4474",
)

go_repository(
    name = "com_github_nkovacs_streamquote",
    importpath = "github.com/nkovacs/streamquote",
    sum = "h1:E2B8qYyeSgv5MXpmzZXRNp8IAQ4vjxIjhpAf5hv/tAg=",
    version = "v0.0.0-20170412213628-49af9bddb229",
)

go_repository(
    name = "com_github_op_go_logging",
    importpath = "github.com/op/go-logging",
    sum = "h1:lDH9UUVJtmYCjyT0CI4q8xvlXPxeZ0gYCVvWbmPlp88=",
    version = "v0.0.0-20160315200505-970db520ece7",
)

go_repository(
    name = "com_github_pkg_errors",
    importpath = "github.com/pkg/errors",
    sum = "h1:WdK/asTD0HN+q6hsWO3/vpuAkAr+tw6aNJNDFFf0+qw=",
    version = "v0.8.0",
)

go_repository(
    name = "com_github_pmezard_go_difflib",
    importpath = "github.com/pmezard/go-difflib",
    sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_sirupsen_logrus",
    importpath = "github.com/sirupsen/logrus",
    sum = "h1:hI/7Q+DtNZ2kINb6qt/lS+IyXnHQe9e90POfeewL/ME=",
    version = "v1.3.0",
)

go_repository(
    name = "com_github_stretchr_objx",
    importpath = "github.com/stretchr/objx",
    sum = "h1:2vfRuCMp5sSVIDSqO8oNnWJq7mPa6KVP3iPIwFBuy8A=",
    version = "v0.1.1",
)

go_repository(
    name = "com_github_stretchr_testify",
    importpath = "github.com/stretchr/testify",
    sum = "h1:TivCn/peBQ7UY8ooIcPgZFpTNSz0Q2U6UrFlUfqbe0Q=",
    version = "v1.3.0",
)

go_repository(
    name = "com_github_valyala_bytebufferpool",
    importpath = "github.com/valyala/bytebufferpool",
    sum = "h1:GqA5TC/0021Y/b9FG4Oi9Mr3q7XYx6KllzawFIhcdPw=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_valyala_fasttemplate",
    importpath = "github.com/valyala/fasttemplate",
    sum = "h1:tY9CJiPnMXf1ERmG2EyK7gNUd+c6RKGD0IfU8WdUSz8=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_weppos_publicsuffix_go",
    importpath = "github.com/weppos/publicsuffix-go",
    sum = "h1:rutRtjBJViU/YjcI5d80t4JAVvDltS6bciJg2K1HrLU=",
    version = "v0.5.0",
)

go_repository(
    name = "com_github_ziutek_mymysql",
    importpath = "github.com/ziutek/mymysql",
    sum = "h1:GB0qdRGsTwQSBVYuVShFBKaXSnSnYYC2d9knnE1LHFs=",
    version = "v1.5.4",
)

go_repository(
    name = "com_github_zmap_rc2",
    importpath = "github.com/zmap/rc2",
    sum = "h1:kKCF7VX/wTmdg2ZjEaqlq99Bjsoiz7vH6sFniF/vI4M=",
    version = "v0.0.0-20131011165748-24b9757f5521",
)

go_repository(
    name = "com_github_zmap_zcertificate",
    importpath = "github.com/zmap/zcertificate",
    sum = "h1:17HHAgFKlLcZsDOjBOUrd5hDihb1ggf+1a5dTbkgkIY=",
    version = "v0.0.0-20180516150559-0e3d58b1bac4",
)

go_repository(
    name = "com_github_zmap_zcrypto",
    importpath = "github.com/zmap/zcrypto",
    sum = "h1:mvOa4+/DXStR4ZXOks/UsjeFdn5O5JpLUtzqk9U8xXw=",
    version = "v0.0.0-20190729165852-9051775e6a2e",
)

go_repository(
    name = "com_github_zmap_zlint",
    importpath = "github.com/zmap/zlint",
    sum = "h1:vxqkjztXSaPVDc8FQCdHTaejm2x747f6yPbnu1h2xkg=",
    version = "v0.0.0-20190806154020-fd021b4cfbeb",
)

go_repository(
    name = "in_gopkg_check_v1",
    importpath = "gopkg.in/check.v1",
    sum = "h1:qIbj1fsPNlZgppZ+VLlY7N33q108Sa+fhmuc+sWQYwY=",
    version = "v1.0.0-20180628173108-788fd7840127",
)

go_repository(
    name = "org_bitbucket_liamstask_goose",
    importpath = "bitbucket.org/liamstask/goose",
    sum = "h1:bkb2NMGo3/Du52wvYj9Whth5KZfMV6d3O0Vbr3nz/UE=",
    version = "v0.0.0-20150115234039-8488cc47d90c",
)

go_repository(
    name = "org_golang_x_lint",
    importpath = "golang.org/x/lint",
    sum = "h1:5hukYrvBGR8/eNkX5mdUezrA6JiaEZDtJb9Ei+1LlBs=",
    version = "v0.0.0-20190930215403-16217165b5de",
)

go_repository(
    name = "org_golang_x_sys",
    importpath = "golang.org/x/sys",
    sum = "h1:+R4KGOnez64A81RvjARKc4UT5/tI9ujCIVX+P5KiHuI=",
    version = "v0.0.0-20190412213103-97732733099d",
)

go_repository(
    name = "org_golang_x_tools",
    importpath = "golang.org/x/tools",
    sum = "h1:/e+gpKk9r3dJobndpTytxS2gOy6m5uvpg+ISQoEcusQ=",
    version = "v0.0.0-20190311212946-11955173bddd",
)

http_archive(
    name = "com_justbuchanan_rules_qt",
    sha256 = "a25924da08346be8f4a0b019bc3cb7a6fdc18d82c35b38b70a0fdc9a118b0846",
    strip_prefix = "bazel_rules_qt-b293b5afd8772ef7de0069ad01a96e373535ee6e",
    urls = [
        "https://github.com/justbuchanan/bazel_rules_qt/archive/b293b5afd8772ef7de0069ad01a96e373535ee6e.tar.gz",
    ],
)

http_archive(
    name = "build_bazel_rules_apple",
    sha256 = "af814a595939cea6c8d3be09fc0fdfb0ff96703f5c21c789b300fd4226298586",
    strip_prefix = "rules_apple-09dbddf9c5d203d1dbb2b8a38f1ad77b3813bf72",
    urls = [
        "https://github.com/bazelbuild/rules_apple/archive/09dbddf9c5d203d1dbb2b8a38f1ad77b3813bf72.tar.gz",
    ],
)

http_archive(
    name = "build_bazel_rules_swift",
    sha256 = "e7dfaf708c47d0fb8919c2b81cb3504bfbc31851fdcde666474e0e22e9f6efd8",
    strip_prefix = "rules_swift-b1610a1b21480851ca37d245795dbfb3b3c7c2fd",
    urls = ["https://github.com/bazelbuild/rules_swift/archive/b1610a1b21480851ca37d245795dbfb3b3c7c2fd.tar.gz"],
)

http_archive(
    name = "build_bazel_apple_support",
    sha256 = "1306d854604344933114b55fc9690239d19f1fe4575d3e09bc93268a5f74751c",
    strip_prefix = "apple_support-647480b45440131b34d9cc8ffbbba39b72acccb2",
    urls = ["https://github.com/bazelbuild/apple_support/archive/647480b45440131b34d9cc8ffbbba39b72acccb2.tar.gz"],
)

load(
    "@build_bazel_rules_apple//apple:repositories.bzl",
    "apple_rules_dependencies",
)

apple_rules_dependencies()

load(
    "@build_bazel_rules_swift//swift:repositories.bzl",
    "swift_rules_dependencies",
)

swift_rules_dependencies()

load(
    "@build_bazel_apple_support//lib:repositories.bzl",
    "apple_support_dependencies",
)

apple_support_dependencies()

# TODO build required qt libs from source
# http_archive(
#     name = "qt",
#     strip_prefix="qtbase-ba3b53cb501a77144aa6259e48a8e0edc3d1481d",
#     build_file = "@com_justbuchanan_rules_qt//:qt.BUILD",
#     urls = [
#         "https://github.com/qt/qtbase/archive/ba3b53cb501a77144aa6259e48a8e0edc3d1481d.tar.gz",
#     ],
# )

new_local_repository(
    name = "qt_linux",
    build_file = "@com_justbuchanan_rules_qt//:qt.BUILD",
    path = "/usr/include/x86_64-linux-gnu/qt5/",
)

new_local_repository(
    name = "qt_mac",
    build_file = "//third_party:qt-mac.BUILD",
    path = "/usr/local/opt/qt/include/",
)

new_local_repository(
    name = "qt_msys",
    build_file = "@com_justbuchanan_rules_qt//:qt.BUILD",
    path = "C:\\msys64\\mingw64\\include\\",
)

go_repository(
    name = "com_github_hashicorp_golang_lru",
    importpath = "github.com/hashicorp/golang-lru",
    sum = "h1:YDjusn29QI/Das2iO9M0BHnIbxPeyuCHsjMW+lJfyTc=",
    version = "v0.5.4",
)

go_repository(
    name = "org_uber_go_zap",
    importpath = "go.uber.org/zap",
    sum = "h1:ZZCA22JRF2gQE5FoNmhmrf7jeJJ2uhqDUNRYKm8dvmM=",
    version = "v1.15.0",
)

go_repository(
    name = "org_uber_go_atomic",
    importpath = "go.uber.org/atomic",
    sum = "h1:Ezj3JGmsOnG1MoRWQkPBsKLe9DwWD9QeXzTRzzldNVk=",
    version = "v1.6.0",
)

go_repository(
    name = "org_uber_go_multierr",
    importpath = "go.uber.org/multierr",
    sum = "h1:KCa4XfM8CWFCpxXRGok+Q0SS/0XBhMDbHHGABQLvD2A=",
    version = "v1.5.0",
)

go_repository(
    name = "io_nhooyr_websocket",
    importpath = "nhooyr.io/websocket",
    sum = "h1:s+C3xAMLwGmlI31Nyn/eAehUlZPwfYZu2JXM621Q5/k=",
    version = "v1.8.6",
)

go_repository(
    name = "com_github_klauspost_compress",
    importpath = "github.com/klauspost/compress",
    sum = "h1:a/y8CglcM7gLGYmlbP/stPE5sR3hbhFRUjCBfd/0B3I=",
    version = "v1.10.10",
)

go_repository(
    name = "com_github_productimon_wasmws",
    importpath = "github.com/productimon/wasmws",
    sum = "h1:GWULhL88aJNhBU9YueXRYfbg5DbxR/CtWXkYFay1rxg=",
    version = "v0.0.0-20200722143936-311c88976439",
)

go_repository(
    name = "com_github_agnivade_levenshtein",
    importpath = "github.com/agnivade/levenshtein",
    sum = "h1:n6qGwyHG61v3ABce1rPVZklEYRT8NFpCMrpZdBUbYGM=",
    version = "v1.1.0",
)
