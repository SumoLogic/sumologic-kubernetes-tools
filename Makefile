BUILD_TAG ?= dev-latest
IMAGE_NAME = kubernetes-tools
DOCKERHUB_REPO_NAME = sumologic
REPO_URL = $(DOCKERHUB_REPO_NAME)/$(IMAGE_NAME)
ECR_URL = public.ecr.aws/a4t4y2n3
ECR_REPO_URL = $(ECR_URL)/$(IMAGE_NAME)

markdownlint: mdl

mdl:
	mdl --style .markdownlint/style.rb .

.PHONY: add-tag
add-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Adding tag ${TAG}"
	@git tag -a ${TAG} -s -m "${TAG}"

.PHONY: push-tag
push-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Pushing tag ${TAG}"
	@git push origin ${TAG}

.PHONY: delete-tag
delete-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Deleting tag ${TAG}"
	@git tag -d ${TAG}

.PHONY: delete-remote-tag
delete-remote-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Deleting remote tag ${TAG}"
	@git push --delete origin ${TAG}

build-image:
	TAG=$(BUILD_TAG) docker buildx bake

build-image-multiplatform:
	TAG=$(BUILD_TAG) docker buildx bake tools-multiplatform 

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
	IMAGE=$(REPO_URL) TAG=$(BUILD_TAG) docker buildx bake tools-multiplatform --push

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
