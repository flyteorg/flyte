# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'LYFT/BOILERPLATE' REPOSITORY:
# 
# TO OPT OUT OF UPDATES, SEE https://github.com/lyft/boilerplate/blob/master/Readme.rst

.PHONY: lint
lint: #lints the package for common code smells
	which golangci-lint || GO111MODULE=on go install github.com/golangci/golangci-lint/cmd/golangci-lint
	# Calling lint twice here is a hack. The first call seem to fail when internally calling `go list...`
	# However, that call seem to have some effects (e.g. https://github.com/golang/go/issues/29452) which, for some
	# reason, allows the subsequent calls to succeed.
	# TODO: Evaluate whether this is still a problem after moving admin dependency system to go modules.
	GO111MODULE=off GL_DEBUG=linters_output,loader,env golangci-lint run --exclude deprecated -v || true
	GO111MODULE=off GL_DEBUG=linters_output,loader,env golangci-lint run --deadline=5m --exclude deprecated -v

# If code is failing goimports linter, this will fix.
# skips 'vendor'
.PHONY: goimports
goimports:
	@boilerplate/lyft/golang_test_targets/goimports

.PHONY: install
install: #download dependencies (including test deps) for the package
	which dep || (curl "https://raw.githubusercontent.com/golang/dep/master/install.sh" | sh)
	dep ensure --vendor-only

.PHONY: test_unit
test_unit:
	go test -cover ./... -race

.PHONY: test_benchmark
test_benchmark:
	go test -bench . ./...

.PHONY: test_unit_cover
test_unit_cover:
	go test ./... -coverprofile /tmp/cover.out -covermode=count; go tool cover -func /tmp/cover.out

.PHONY: test_unit_visual
test_unit_visual:
	go test ./... -coverprofile /tmp/cover.out -covermode=count; go tool cover -html=/tmp/cover.out
