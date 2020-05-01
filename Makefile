export REPOSITORY=flyteconsole
include boilerplate/lyft/docker_build/Makefile

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

.PHONY: install
install: #installs dependencies
	yarn

.PHONY: install_ci
install_ci: install
	yarn add codecov

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
	yarn run test-coverage && yarn run codecov
