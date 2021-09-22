export REPOSITORY=flyteconsole
include boilerplate/flyte/docker_build/Makefile

.PHONY: update_boilerplate
update_boilerplate:
	@curl https://raw.githubusercontent.com/flyteorg/boilerplate/master/boilerplate/update.sh -o boilerplate/update.sh
	@boilerplate/update.sh

.PHONY: install
install: #installs dependencies
	yarn

.PHONY: lint
lint: #lints the package for common code smells
	yarn run lint

.PHONY: build_prod
build_prod:
	yarn run clean
	BASE_URL=/console yarn run build:prod

# test_unit runs all unit tests
.PHONY: test_unit
test_unit:
	yarn test

# server starts the service in development mode
.PHONY: server
server:
	yarn start

.PHONY: clean
clean:
	yarn run clean

# test_unit_codecov runs unit tests with code coverage turned on and
# submits the coverage to codecov.io
.PHONY: test_unit_codecov
test_unit_codecov:
	yarn run test-coverage

.PHONY: generate_ssl
generate_ssl:
	./script/generate_ssl.sh
