group "default" {
  targets = ["actions"]
}

target "actions" {
  dockerfile = "docker/Dockerfile.actions"
  context = "."
  platforms = ["linux/arm64"]
  tags = ["flyte-actions-service:latest"]
}
