variable "IMAGE" {
    default = "kubernetes-tools"
}

variable "TAG" {
    default = "dev-latest"
}

variable "CACHE_IMAGE" {
    default = "sumologic/kubernetes-tools"
}

variable "BUILD_GO_CACHE_TAG" {
    default = "go-build-cache"
}

variable "BUILD_RUST_CACHE_TAG" {
    default = "rust-build-cache"
}

variable "TOOLS_CACHE_TAG" {
    default = "tools-build-cache"
}

target "multiplatform" {
    platforms = ["linux/amd64", "linux/arm64"]
    output = ["type=image"]
}

target "default" {
    dockerfile = "Dockerfile"
    tags = ["${IMAGE}:${TAG}"]
    cache-from = [
        "${CACHE_IMAGE}:${BUILD_GO_CACHE_TAG}",
        "${CACHE_IMAGE}:${BUILD_RUST_CACHE_TAG}",
        "${CACHE_IMAGE}:${TOOLS_CACHE_TAG}",
    ]
    output = ["type=docker"]
    platforms = ["linux/amd64"]
}

target "kubectl" {
    dockerfile = "Dockerfile.kubectl"
    tags = ["${IMAGE}-kubectl:${TAG}"]
    output = ["type=docker"]
    platforms = ["linux/amd64"]
}

target "tools-multiplatform" {
    inherits = ["default", "multiplatform"]
}

target "kubeclt-multiplatform" {
    inherits = ["kubectl", "multiplatform"]
}

group "cache" {
    targets = ["rust-cache", "go-cache", "tools-cache"]
}

target "rust-cache" {
    dockerfile = "Dockerfile"
    cache-from = [
        "${CACHE_IMAGE}:${BUILD_RUST_CACHE_TAG}",
    ]
    cache-to = ["${CACHE_IMAGE}:${BUILD_RUST_CACHE_TAG}"]
    target = "rust-builder"
}

target "go-cache" {
    dockerfile = "Dockerfile"
    cache-from = [
        "${CACHE_IMAGE}:${BUILD_GO_CACHE_TAG}",
    ]
    cache-to = ["${CACHE_IMAGE}:${BUILD_GO_CACHE_TAG}"]
    target = "go-builder"
}

target "tools-cache" {
    inherits = ["default"]
    cache-to = ["${CACHE_IMAGE}:${TOOLS_CACHE_TAG}"]
}

group "cache-multiplatform" {
    targets = ["rust-cache-multiplatform", "go-cache-multiplatform", "tools-cache-multiplatform"]
}

target "rust-cache-multiplatform" {
    inherits = ["rust-cache", "multiplatform"]
}

target "go-cache-multiplatform" {
    inherits = ["go-cache", "multiplatform"]
}

target "tools-cache-multiplatform" {
    inherits = ["tools-cache", "multiplatform"]
}
