export REPOSITORY=flyteplugins
include boilerplate/lyft/golang_test_targets/Makefile

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

generate: download_tooling
	@go generate ./...

clean:
	rm -rf bin
