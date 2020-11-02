#!/bin/bash

VERSION="${TRAVIS_TAG:-0.0.0}"
VERSION="${VERSION#v}"
: "${DOCKER_IMAGE:=sumologic/kubernetes-tools}"
: "${DOCKER_USERNAME:=sumodocker}"

readonly DOCKER_PASSWORD=${DOCKER_PASSWORD:-}
readonly GITHUB_TOKEN=${GITHUB_TOKEN:-}
readonly TRAVIS_BRANCH=${TRAVIS_BRANCH:-}
readonly TRAVIS_EVENT_TYPE=${TRAVIS_EVENT_TYPE:-}
readonly TRAVIS_PULL_REQUEST_BRANCH=${TRAVIS_PULL_REQUEST_BRANCH:-}

err_report() {
    echo "Script error on line $1"
    exit 1
}
trap 'err_report $LINENO' ERR

# Functions

function push_docker_image() {
  local version="$1"

  echo "Tagging docker image ${DOCKER_IMAGE}:local with ${DOCKER_IMAGE}:${version}..."
  docker tag "${DOCKER_IMAGE}:local" "${DOCKER_IMAGE}:${version}"
  echo "Pushing docker image ${DOCKER_IMAGE}:${version}..."
  echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin
  docker push "${DOCKER_IMAGE}:${version}"
}

function set_up_github() {
  if [[ -z "${GITHUB_TOKEN}" ]]; then
    echo "GITHUB_TOKEN not provided - skipping git setup"
    return
  fi

  git config --global user.email "travis@travis-ci.org"
  git config --global user.name "Travis CI"
  git remote add origin-repo "https://${GITHUB_TOKEN}@github.com/SumoLogic/sumologic-kubernetes-tools.git" > /dev/null 2>&1
  git fetch --unshallow origin-repo
  git checkout "${TRAVIS_PULL_REQUEST_BRANCH}"
}

function build_docker_image() {
  echo "Building docker image with ${DOCKER_IMAGE}:local in $(pwd)..."
  docker build . -f deploy/docker/Dockerfile -t "${DOCKER_IMAGE}:local" --no-cache
}

function test_docker_image() {
  local apps="stress-tester k8s-api-test receiver-mock check fix-log-symlinks tools-usage template template-dependency template-prometheus-mixin"

  echo
  echo "Running docker image tests..."

  for app in ${apps}; do
    echo "Testing ${app}..."
    docker run --rm "${DOCKER_IMAGE}:local" "${app}" --help >/dev/null
  done

  echo "Testing docker image's CMD..."
  docker run --rm "${DOCKER_IMAGE}:local" >/dev/null

  echo "Docker image tests OK"
}

function push_docker_images() {
  if [[ -z "${DOCKER_PASSWORD}" ]]; then
    echo "Skip Docker image push (DOCKER_PASSWORD was not provided)"
    return
  fi

  if [[ -n "${TRAVIS_TAG}" ]]; then
    push_docker_image "${VERSION}"

    # if the tag is a GA and not a prerelease then retag the latest to it
    if [[ "${TRAVIS_TAG}" =~ v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
      push_docker_image "latest"
    fi
  elif [[ "${TRAVIS_BRANCH}" == "master" || "${TRAVIS_BRANCH}" =~ ^release-v[0-9]+\.[0-9]+$ ]] && [[ "${TRAVIS_EVENT_TYPE}" == "push" ]]; then
    dev_build_tag="$(git describe --tags --always)"
    dev_build_tag="${dev_build_tag#v}"
    push_docker_image "${dev_build_tag}"
    push_docker_image "${TRAVIS_BRANCH}"
  fi
}

# Main

echo "Starting build process in '$(pwd)' with version tag: ${VERSION}"
set_up_github
build_docker_image
test_docker_image
push_docker_images
echo "Done with build.sh"
