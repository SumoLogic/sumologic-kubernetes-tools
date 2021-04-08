#!/usr/bin/env bash

set -e

function test_image() {
  local apps="
  stress-tester
  k8s-api-test
  receiver-mock
  check
  fix-log-symlinks
  tools-usage
  template
  template-dependency
  template-prometheus-mixin
  logs-generator
  "
  readonly apps
  local tag="${1}"
  readonly tag

  echo "Testing docker image ${tag}"

  for app in ${apps}; do
    echo "Testing ${app}..."
    docker run --rm "${tag}" "${app}" --help >/dev/null
  done

  echo "Testing docker image's CMD..."
  docker run --rm "${tag}" >/dev/null

  echo "Docker image tests OK"
}

function check_if_image_available(){
  local tag="${1}"
  readonly tag

  if [[ -z "$(docker images -q "${tag}")" ]]; then
    echo "Docker image ${tag} unavailable. Build it or tag an existing image"
    exit 1
  fi
}

check_if_image_available "$@"
test_image "$@"
