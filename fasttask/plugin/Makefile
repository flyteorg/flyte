export REPOSITORY=fasttask/plugin

include boilerplate/flyte/golang_test_targets/Makefile

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

.PHONY: format
format:
	@GL_DEBUG=linters_output,env golangci-lint run --deadline=5m -v --fix

.PHONY: go-tidy
go-tidy:
	go mod tidy
