BUILD_TAG ?= dev-latest
IMAGE_NAME = kubernetes-tools
DOCKERHUB_REPO_NAME = sumologic
REPO_URL = $(DOCKERHUB_REPO_NAME)/$(IMAGE_NAME)
ECR_URL = public.ecr.aws/a4t4y2n3
ECR_REPO_URL = $(ECR_URL)/$(IMAGE_NAME)

CHMOD_IMAGE_NAME = kubernetes-chmod
CHMOD_DOCKERFILE = Dockerfile.chmod
CHMOD_REPO_URL = $(DOCKERHUB_REPO_NAME)/$(CHMOD_IMAGE_NAME)
CHMOD_ECR_REPO_URL = $(ECR_URL)/$(CHMOD_IMAGE_NAME)

markdownlint: mdl

mdl:
	mdl --style .markdownlint/style.rb .

build-image:
	TAG=$(BUILD_TAG) docker buildx bake

build-image-multiplatform:
	TAG=$(BUILD_TAG) docker buildx bake image-multiplatform

tag-release-image-with-latest:
	make push-image BUILD_TAG=latest

tag-release-image-with-latest-ecr:
	make tag-release-image-with-latest REPO_URL=$(ECR_REPO_URL)

test-image:
	./scripts/test-image.sh "$(IMAGE_NAME):$(BUILD_TAG)"

push-image-cache:
	# only push cache to Dockerhub as ECR doesn't support it yet
    	# https://github.com/aws/containers-roadmap/issues/876
	docker buildx bake cache-multiplatform

push-image:
	IMAGE=$(REPO_URL) TAG=$(BUILD_TAG) docker buildx bake image-multiplatform --push

push-image-ecr:
	make push-image REPO_URL=$(ECR_REPO_URL)

login:
	echo "${DOCKER_PASSWORD}" | docker login -u sumodocker --password-stdin

login-ecr:
	aws ecr-public get-login-password --region us-east-1 \
	| docker login --username AWS --password-stdin $(ECR_URL)

build-update-collection-v3:
	make build -C src/go/cmd/update-collection-v3

test-update-collection-v3:
	make test -C src/go/cmd/update-collection-v3

# chmod image

build-chmod-image-multiplatform:
	TAG=$(BUILD_TAG) DOCKERFILE=$(CHMOD_DOCKERFILE) IMAGE=$(CHMOD_IMAGE_NAME) docker buildx bake image-multiplatform

build-chmod-image:
	TAG=$(BUILD_TAG) DOCKERFILE=$(CHMOD_DOCKERFILE) IMAGE=$(CHMOD_IMAGE_NAME) docker buildx bake

push-chmod-image:
	TAG=$(BUILD_TAG) DOCKERFILE=$(CHMOD_DOCKERFILE) IMAGE=$(CHMOD_REPO_URL) docker buildx bake image-multiplatform --push

tag-release-chmod-image-with-latest:
	make push-image BUILD_TAG=latest

push-chmod-image-ecr:
	make push-image REPO_URL=$(CHMOD_ECR_REPO_URL)
