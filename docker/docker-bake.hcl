variable "IMAGE_REPO_PREFIX" {
  default = "ghcr.io/flyteorg/flyte"
}

variable "IMAGE_SHA_SHORT_TAG" {
  default = "latest"
}

variable "GO_VERSION" {
  default = "1.24"
}

group "default" {
  targets = [
    "actions",
    "cache_service",
    "dataproxy",
    "events",
    "executor",
    "runs",
  ]
}

target "_common" {
  context = "."
  platforms = ["linux/arm64"]
  args = {
    GO_VERSION = "${GO_VERSION}"
  }
}

target "actions" {
  inherits   = ["_common"]
  dockerfile = "docker/Dockerfile.actions"
  tags = ["${IMAGE_REPO_PREFIX}/actions:${IMAGE_SHA_SHORT_TAG}"]
}

target "cache_service" {
  inherits   = ["_common"]
  dockerfile = "docker/Dockerfile.cache_service"
  tags = ["${IMAGE_REPO_PREFIX}/cache_service:${IMAGE_SHA_SHORT_TAG}"]
}

target "dataproxy" {
  inherits   = ["_common"]
  dockerfile = "docker/Dockerfile.dataproxy"
  tags = ["${IMAGE_REPO_PREFIX}/dataproxy:${IMAGE_SHA_SHORT_TAG}"]
}

target "events" {
  inherits   = ["_common"]
  dockerfile = "docker/Dockerfile.events"
  tags = ["${IMAGE_REPO_PREFIX}/events:${IMAGE_SHA_SHORT_TAG}"]
}

target "executor" {
  inherits   = ["_common"]
  dockerfile = "docker/Dockerfile.executor"
  tags = ["${IMAGE_REPO_PREFIX}/executor:${IMAGE_SHA_SHORT_TAG}"]
}

target "runs" {
  inherits   = ["_common"]
  dockerfile = "docker/Dockerfile.runs"
  tags = ["${IMAGE_REPO_PREFIX}/runs:${IMAGE_SHA_SHORT_TAG}"]
}
