BUILD_TAG ?= dev-latest
BUILD_GO_CACHE_TAG = go-build-cache
BUILD_RUST_CACHE_TAG = rust-build-cache
IMAGE_NAME = kubernetes-tools
DOCKERHUB_REPO_NAME = sumologic
REPO_URL = $(DOCKERHUB_REPO_NAME)/$(IMAGE_NAME)

markdownlint: mdl

mdl:
	mdl --style .markdownlint/style.rb .

build-image:
	DOCKER_BUILDKIT=1 docker build \
		--build-arg BUILDKIT_INLINE_CACHE=1 \
		--cache-from $(REPO_URL):$(BUILD_GO_CACHE_TAG) \
		--target go-builder \
		--tag $(IMAGE_NAME):$(BUILD_GO_CACHE_TAG) \
		.

	DOCKER_BUILDKIT=1 docker build \
		--build-arg BUILDKIT_INLINE_CACHE=1 \
		--cache-from $(REPO_URL):$(BUILD_RUST_CACHE_TAG) \
		--target rust-builder \
		--tag $(IMAGE_NAME):$(BUILD_RUST_CACHE_TAG) \
		.

	DOCKER_BUILDKIT=1 docker build \
		--build-arg BUILDKIT_INLINE_CACHE=1 \
		--cache-from $(REPO_URL):$(BUILD_GO_CACHE_TAG) \
		--cache-from $(REPO_URL):$(BUILD_RUST_CACHE_TAG) \
		--cache-from $(REPO_URL):dev-latest \
		--tag $(IMAGE_NAME):$(BUILD_TAG) \
		.

test-image:
	./scripts/test-image.sh "$(IMAGE_NAME):$(BUILD_TAG)"

push-image-cache:
	docker tag $(IMAGE_NAME):$(BUILD_GO_CACHE_TAG) $(REPO_URL):$(BUILD_GO_CACHE_TAG)
	docker push $(REPO_URL):$(BUILD_GO_CACHE_TAG)
	docker tag $(IMAGE_NAME):$(BUILD_RUST_CACHE_TAG) $(REPO_URL):$(BUILD_RUST_CACHE_TAG)
	docker push $(REPO_URL):$(BUILD_RUST_CACHE_TAG)
	docker tag $(IMAGE_NAME):$(BUILD_TAG) $(REPO_URL):dev-latest
	docker push $(REPO_URL):dev-latest

push-image:
	docker tag $(IMAGE_NAME):$(BUILD_TAG) $(REPO_URL):$(BUILD_TAG)
	docker push $(REPO_URL):$(BUILD_TAG)

login:
	echo "${DOCKER_PASSWORD}" | docker login -u sumodocker --password-stdin
