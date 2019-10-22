.PHONY: lint
lint: #lints the package for common code smells
	which golangci-lint || GO111MODULE=on go install github.com/golangci/golangci-lint/cmd/golangci-lint
	golangci-lint run

# If code is failing goimports linter, this will fix.
# skips 'vendor'
.PHONY: goimports
goimports:
	@boilerplate/lyft/golang_test_targets/goimports

.PHONY: install
install: #download dependencies (including test deps) for the package
	which dep || (curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh)
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
