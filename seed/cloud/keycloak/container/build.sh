#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

container_cli=${CONTAINER_CLI:-"docker"}

bazel build //seed/cloud/keycloak:keycloak_tar_gz
cp -f ./bazel-bin/seed/cloud/keycloak/keycloak.tar.gz ./seed/cloud/keycloak/container/keycloak.tar.gz

bazel build @keycloak_maven//:com_kohlschutter_junixsocket_junixsocket_common_2_10_1
cp -f \
  ./bazel-bin/external/rules_jvm_external++_maven+++keycloak_maven+maven/v1/https/repo1.maven.org/maven2/com/kohlschutter/junixsocket/junixsocket-common/2.10.1/processed_junixsocket-common-2.10.1.jar \
  ./seed/cloud/keycloak/container/junixsocket-common.jar

bazel build @keycloak_maven//:com_kohlschutter_junixsocket_junixsocket_native_common_2_10_1
cp -f \
  ./bazel-bin/external/rules_jvm_external++_maven+++keycloak_maven+maven/v1/https/repo1.maven.org/maven2/com/kohlschutter/junixsocket/junixsocket-native-common/2.10.1/processed_junixsocket-native-common-2.10.1.jar \
  ./seed/cloud/keycloak/container/junixsocket-native-common.jar

cd ./seed/cloud/keycloak/container/
"${container_cli}" build -t ghcr.io/ndscm/seed-cloud-keycloak-container:latest .
