# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'FLYTEORG/BOILERPLATE' REPOSITORY:
#
# TO OPT OUT OF UPDATES, SEE https://github.com/flyteorg/boilerplate/blob/master/Readme.rst

LINT_FLAGS ?=

.PHONY: download_tooling
download_tooling: #download dependencies (including test deps) for the package
	@../boilerplate/flyte/golang_test_targets/download_tooling.sh

.PHONY: generate
generate: download_tooling #generate go code
	@../boilerplate/flyte/golang_test_targets/go-gen.sh

.PHONY: lint
lint: download_tooling #lints the package for common code smells
	GL_DEBUG=linters_output,env golangci-lint run $(LINT_FLAGS) --deadline=5m --exclude deprecated -v

.PHONY: lint-fix
lint-fix: LINT_FLAGS=--fix
lint-fix: lint

.PHONY: mod_download
mod_download: #download dependencies (including test deps) for the package
	go mod download

.PHONY: install
install: download_tooling mod_download

.PHONY: show
show:
	go list -m all

.PHONY: test_unit
test_unit:
	go test -cover ./... -race

.PHONY: test_benchmark
test_benchmark:
	go test -bench . ./...

.PHONY: test_unit_cover
test_unit_cover:
	go test ./... -coverprofile /tmp/cover.out -covermode=count
	go tool cover -func /tmp/cover.out

.PHONY: test_unit_visual
test_unit_visual:
	go test ./... -coverprofile /tmp/cover.out -covermode=count
	go tool cover -html=/tmp/cover.out

.PHONY: test_unit_codecov
test_unit_codecov:
	go test ./... -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: go-tidy
go-tidy:
	go mod tidy
