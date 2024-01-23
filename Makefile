BUILD_TAG ?= dev-latest
IMAGE_NAME = kubernetes-tools
DOCKERHUB_REPO_NAME = sumologic
REPO_URL = $(DOCKERHUB_REPO_NAME)/$(IMAGE_NAME)
ECR_URL = public.ecr.aws/a4t4y2n3
ECR_REPO_URL = $(ECR_URL)/$(IMAGE_NAME)

SUMOLOGIC_MOCK_IMAGE_NAME = sumologic-mock
SUMOLOGIC_MOCK_REPO_URL = $(DOCKERHUB_REPO_NAME)/$(SUMOLOGIC_MOCK_IMAGE_NAME)
SUMOLOGIC_MOCK_ECR_REPO_URL = $(ECR_URL)/$(SUMOLOGIC_MOCK_IMAGE_NAME)

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

build-image: build-image-tools build-image-kubectl

build-image-tools:
	TAG=$(BUILD_TAG) docker buildx bake

build-image-kubectl:
	TAG=$(BUILD_TAG) docker buildx bake kubectl

build-image-sumologic-mock:
	TAG=$(BUILD_TAG) docker buildx bake sumologic-mock

build-image-multiplatform: build-image-multiplatform-tools build-image-multiplatform-kubectl

build-image-multiplatform-tools:
	TAG=$(BUILD_TAG) docker buildx bake tools-multiplatform

build-image-multiplatform-kubectl:
	TAG=$(BUILD_TAG) docker buildx bake kubectl-multiplatform

build-image-multiplatform-sumologic-mock:
	TAG=$(BUILD_TAG) docker buildx bake sumologic-mock-multiplatform

tag-release-image-with-latest-tools:
	make push-image-tools BUILD_TAG=latest

tag-release-image-with-latest-kubectl:
	make push-image-kubectl BUILD_TAG=latest

tag-release-image-with-latest-ecr-tools:
	make tag-release-image-with-latest-tools REPO_URL=$(ECR_REPO_URL)

tag-release-image-with-latest-ecr-kubectl:
	make tag-release-image-with-latest-kubectl REPO_URL=$(ECR_REPO_URL)

test-image:
	./scripts/test-image.sh "$(IMAGE_NAME):$(BUILD_TAG)"

push-image-cache-tools:
	# only push cache to Dockerhub as ECR doesn't support it yet
    	# https://github.com/aws/containers-roadmap/issues/876
	docker buildx bake cache-multiplatform

push-image-tools:
	IMAGE=$(REPO_URL) TAG=$(BUILD_TAG) docker buildx bake tools-multiplatform --push

push-image-kubectl:
	IMAGE=$(REPO_URL) TAG=$(BUILD_TAG) docker buildx bake kubectl-multiplatform --push

push-image-sumologic-mock:
	IMAGE=$(SUMOLOGIC_MOCK_REPO_URL) TAG=$(BUILD_TAG) docker buildx bake sumologic-mock-multiplatform --push

push-image-ecr-tools:
	make push-image-tools REPO_URL=$(ECR_REPO_URL)

push-image-ecr-kubectl:
	make push-image-kubectl REPO_URL=$(ECR_REPO_URL)

push-image-ecr-sumologic-mock:
	make push-image-sumologic-mock REPO_URL=$(SUMOLOGIC_MOCK_ECR_REPO_URL)

login:
	echo "${DOCKER_PASSWORD}" | docker login -u sumodocker --password-stdin

login-ecr:
	aws ecr-public get-login-password --region us-east-1 \
	| docker login --username AWS --password-stdin $(ECR_URL)

build-update-collection-v3:
	make build -C src/go/cmd/update-collection-v3

test-update-collection-v3:
	make test -C src/go/cmd/update-collection-v3
