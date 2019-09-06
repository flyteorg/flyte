# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'LYFT/BOILERPLATE' REPOSITORY:
# 
# TO OPT OUT OF UPDATES, SEE https://github.com/lyft/boilerplate/blob/master/Readme.rst

DEP_SHA=1f7c19e5f52f49ffb9f956f64c010be14683468b

.PHONY: lint
lint: #lints the package for common code smells
	which golangci-lint || curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b $$GOPATH/bin v1.16.0
	golangci-lint run --exclude deprecated

# If code is failing goimports linter, this will fix.
# skips 'vendor'
.PHONY: goimports
goimports:
	@boilerplate/lyft/golang_test_targets/goimports

.PHONY: install
install: #download dependencies (including test deps) for the package
	which dep || (curl "https://raw.githubusercontent.com/golang/dep/${DEP_SHA}/install.sh" | sh)
	dep ensure

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
