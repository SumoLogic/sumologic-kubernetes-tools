# Releasing

- [How to release](#how-to-release)
- [Create and push Git tag](#create-and-push-git-tag)
- [Publish GitHub release](#publish-github-release)

## How to release

### Create and push Git tag

In order to release a new version of Sumologic Kubernetes Tools you'd export `TAG` env variable and create a tag and push it.

This can be done using `add-tag` and `push-tag` `make` targets which will handle
that for you.

```shell
export TAG=v3.15.0
make add-tag push-tag
```

#### Remove tag in case of a failed release job

Pushing a new version tag to GitHub starts the [release build](../.github/workflows/release_builds.yml) jobs.

If one of these jobs fails for whatever reason (real world example: failing to notarize the MacOS binary),
you might need to remove the created tags, perhaps change something, and create the tags again.

To delete the tags both locally and remotely, run the following commands:

```shell
export TAG=v3.15.0
make delete-tag delete-remote-tag
```

### Publish GitHub release

The GitHub release is created as draft by the [create-release](../.github/workflows/release_builds.yml) GitHub Action.

After the release draft is created, go to [GitHub releases](https://github.com/SumoLogic/sumologic-kubernetes-tools/releases),
edit the release draft and fill in missing information.

After verifying that the release text and all links are good, publish the release.
