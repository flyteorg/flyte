.PHONY: build_images
build_images:
	ls per-language | xargs -n 1 -I {} docker build ./per-language/{} -f ./per-language/{}/Dockerfile --tag "rawcontainers-{}:v1"
