#!/bin/bash

VERSION="${TRAVIS_TAG:-0.0.0}"
VERSION="${VERSION#v}"
: "${DOCKER_TAG:=sumologic/kubernetes-tools}"
: "${DOCKER_USERNAME:=sumodocker}"
DOCKER_TAGS="https://registry.hub.docker.com/v1/repositories/sumologic/kubernetes-tools/tags"

echo "Starting build process in: $(pwd) with version tag: ${VERSION}"
err_report() {
    echo "Script error on line $1"
    exit 1
}
trap 'err_report $LINENO' ERR

# Set up Github
if [ -n "$GITHUB_TOKEN" ]; then
  git config --global user.email "travis@travis-ci.org"
  git config --global user.name "Travis CI"
  git remote add origin-repo https://${GITHUB_TOKEN}@github.com/SumoLogic/sumologic-kubernetes-tools.git > /dev/null 2>&1
  git fetch --unshallow origin-repo
  git checkout $TRAVIS_PULL_REQUEST_BRANCH
fi

echo "Building docker image with $DOCKER_TAG:local in $(pwd)..."
docker build . -f deploy/docker/Dockerfile -t $DOCKER_TAG:local --no-cache

function push_docker_image() {
  local version="$1"

  echo "Tagging docker image $DOCKER_TAG:local with $DOCKER_TAG:$version..."
  docker tag $DOCKER_TAG:local $DOCKER_TAG:$version
  echo "Pushing docker image $DOCKER_TAG:$version..."
  echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
  docker push $DOCKER_TAG:$version
}

if [ -n "$DOCKER_PASSWORD" ] && [ -n "$TRAVIS_TAG" ]; then
  push_docker_image "$VERSION"

  # if the tag is a GA and not a prerelease then retag the latest to it
  if [[ "$TRAVIS_TAG" =~ v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	  push_docker_image "latest"
  fi

elif [ -n "$DOCKER_PASSWORD" ] && [[ "$TRAVIS_BRANCH" == "master" || "$TRAVIS_BRANCH" =~ ^release-v[0-9]+\.[0-9]+$ ]] && [ "$TRAVIS_EVENT_TYPE" == "push" ]; then
  dev_build_tag=$(git describe --tags --always)
  dev_build_tag=${dev_build_tag#v}
  push_docker_image "$dev_build_tag"
  push_docker_image "$TRAVIS_BRANCH"

else
  echo "Skip Docker pushing"
fi

echo "DONE"
