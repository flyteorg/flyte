# Makefile contains only things for building wheels currently.

# Build both architecture-specific builder images
build-builders: build-arm64 build-amd64

# Build the arm64 builder image
build-arm64:
	docker buildx build --platform linux/arm64 --build-arg ARCH=aarch64 -f Dockerfile.maturinx -t wheel-builder:arm64 .

# Build the amd64 builder image (emulated on arm64 host)
build-amd64:
	docker buildx build --platform linux/amd64 --build-arg ARCH=x86_64 -f Dockerfile.maturinx -t wheel-builder:amd64 .

DIST_DIRS := dist

dist-dirs:
	mkdir -p $(DIST_DIRS) $(CARGO_CACHE_DIR)

define BUILD_WHEELS_RECIPE
docker run --rm \
  $(if $(SETUPTOOLS_SCM_PRETEND_VERSION),-e SETUPTOOLS_SCM_PRETEND_VERSION="$(SETUPTOOLS_SCM_PRETEND_VERSION)") \
  -v $(PWD):/code/fasttask/worker-v2 \
  -v $(PWD)/../../.git:/code/.git \
  wheel-builder:$(1) /bin/bash -c "\
    cd /code/fasttask/worker-v2; \
    /opt/python/cp310-cp310/bin/python -m build --wheel --outdir /code/fasttask/worker-v2/dist \
  "
endef

clean_dist:
	rm -rf $(DIST_DIRS)/*whl

# Build wheels for arm64 (depends on dist-dirs)
build-wheel-arm64: dist-dirs
	$(call BUILD_WHEELS_RECIPE,arm64)

# Build wheels for amd64 (depends on dist-dirs)
build-wheel-amd64: dist-dirs
	$(call BUILD_WHEELS_RECIPE,amd64)

# This is for Mac users, since the other targets won't build macos wheels (only local arch so probably arm64)
build-wheel-local: dist-dirs
	python -m build --wheel --outdir dist

# Only linux, if you want the macos wheel, use build-wheel-local
build-wheels: clean_dist build-wheel-arm64 build-wheel-amd64

# Format Rust code
fmt:
	cargo fmt
