# Common Go Makefile for Flyte services
# Include this file in service-specific Makefiles and set the required variables:
#   SERVICE_NAME - Name of the service (e.g., "manager", "runs", "queue")
#   CMD_PATH     - Path to main.go relative to service directory (default: cmd/main.go)
#   BIN_DIR      - Directory for binaries (default: bin)

# Set defaults
CMD_PATH ?= cmd/main.go
BIN_DIR ?= bin
BIN_NAME ?= $(SERVICE_NAME)
GOLANGCI_LINT_VERSION ?= v2.4.0
LOCALBIN ?= $(shell pwd)/bin

# ANSI color codes for prettier output
COLOR_RESET   := \033[0m
COLOR_BOLD    := \033[1m
COLOR_GREEN   := \033[32m
COLOR_YELLOW  := \033[33m
COLOR_BLUE    := \033[36m

##@ General

.PHONY: help
help: ## Display this help message
	@echo ''
	@echo '$(COLOR_BOLD)Usage:$(COLOR_RESET)'
	@echo '  make $(COLOR_BLUE)<target>$(COLOR_RESET)'
	@echo ''
	@echo '$(COLOR_BOLD)Available targets:$(COLOR_RESET)'
	@awk 'BEGIN {FS = ":.*##"; printf ""} \
		/^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(COLOR_BLUE)%-20s$(COLOR_RESET) %s\n", $$1, $$2 } \
		/^##@/ { printf "\n$(COLOR_BOLD)%s$(COLOR_RESET)\n", substr($$0, 5) } ' \
		$(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt to format code
	@echo "$(COLOR_GREEN)Formatting Go code...$(COLOR_RESET)"
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet to examine code
	@echo "$(COLOR_GREEN)Running go vet...$(COLOR_RESET)"
	@go vet ./...

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter (skips *_test.go files)
	@echo "$(COLOR_GREEN)Running golangci-lint (skipping tests)...$(COLOR_RESET)"
	@$(GOLANGCI_LINT) run --tests=false

.PHONY: lint-all
lint-all: golangci-lint ## Run golangci-lint linter (includes all files)
	@echo "$(COLOR_GREEN)Running golangci-lint (including tests)...$(COLOR_RESET)"
	@$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint and automatically fix issues (skips *_test.go files)
	@echo "$(COLOR_GREEN)Running golangci-lint with auto-fix (skipping tests)...$(COLOR_RESET)"
	@$(GOLANGCI_LINT) run --fix --tests=false

.PHONY: tidy
tidy: ## Run go mod tidy to clean up dependencies
	@echo "$(COLOR_GREEN)Tidying Go modules...$(COLOR_RESET)"
	@go mod tidy

.PHONY: verify
verify: fmt vet ## Run fmt and vet
	@echo "$(COLOR_GREEN)Code verification complete$(COLOR_RESET)"

##@ Build

.PHONY: build
build: verify ## Build the service binary
	@echo "$(COLOR_GREEN)Building $(SERVICE_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(BIN_NAME) $(CMD_PATH)
	@echo "$(COLOR_GREEN)Built: $(BIN_DIR)/$(BIN_NAME)$(COLOR_RESET)"

.PHONY: build-fast
build-fast: ## Build without verification (faster)
	@echo "$(COLOR_GREEN)Building $(SERVICE_NAME) (fast mode)...$(COLOR_RESET)"
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(BIN_NAME) $(CMD_PATH)
	@echo "$(COLOR_GREEN)Built: $(BIN_DIR)/$(BIN_NAME)$(COLOR_RESET)"

.PHONY: install
install: build ## Build and install binary to $GOPATH/bin
	@echo "$(COLOR_GREEN)Installing $(SERVICE_NAME)...$(COLOR_RESET)"
	@go install $(CMD_PATH)

##@ Testing

.PHONY: test
test: ## Run unit tests
	@echo "$(COLOR_GREEN)Running tests...$(COLOR_RESET)"
	@go test -v -race -cover ./...

.PHONY: test-short
test-short: ## Run unit tests (short mode)
	@echo "$(COLOR_GREEN)Running tests (short mode)...$(COLOR_RESET)"
	@go test -short -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(COLOR_GREEN)Running tests with coverage...$(COLOR_RESET)"
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)Coverage report generated: coverage.html$(COLOR_RESET)"

.PHONY: bench
bench: ## Run benchmarks
	@echo "$(COLOR_GREEN)Running benchmarks...$(COLOR_RESET)"
	@go test -bench=. -benchmem ./...

##@ Running

.PHONY: run
run: build-fast ## Build and run the service
	@echo "$(COLOR_GREEN)Running $(SERVICE_NAME)...$(COLOR_RESET)"
	@./$(BIN_DIR)/$(BIN_NAME) $(RUN_ARGS)

.PHONY: run-dev
run-dev: ## Run service directly with go run (no build)
	@echo "$(COLOR_GREEN)Running $(SERVICE_NAME) (dev mode)...$(COLOR_RESET)"
	@go run $(CMD_PATH) $(RUN_ARGS)

##@ Cleanup

.PHONY: clean
clean: ## Remove build artifacts
	@echo "$(COLOR_YELLOW)Cleaning build artifacts...$(COLOR_RESET)"
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)Clean complete$(COLOR_RESET)"

.PHONY: clean-all
clean-all: clean ## Remove build artifacts and installed tools
	@echo "$(COLOR_YELLOW)Cleaning all artifacts and tools...$(COLOR_RESET)"
	@rm -rf $(LOCALBIN)
	@echo "$(COLOR_GREEN)Clean complete$(COLOR_RESET)"

##@ Dependencies

.PHONY: deps
deps: ## Download dependencies
	@echo "$(COLOR_GREEN)Downloading dependencies...$(COLOR_RESET)"
	@go mod download

.PHONY: deps-upgrade
deps-upgrade: ## Upgrade all dependencies
	@echo "$(COLOR_GREEN)Upgrading dependencies...$(COLOR_RESET)"
	@go get -u ./...
	@go mod tidy

##@ Tools

GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary
$(GOLANGCI_LINT): $(LOCALBIN)
	@echo "$(COLOR_GREEN)Installing golangci-lint $(GOLANGCI_LINT_VERSION)...$(COLOR_RESET)"
	@$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

$(LOCALBIN):
	@mkdir -p $(LOCALBIN)

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "  Downloading $${package}" ;\
rm -f $(1) ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $$(basename $(1)-$(3)) $(1)
endef

##@ Info

.PHONY: info
info: ## Display service information
	@echo "$(COLOR_BOLD)Service Information:$(COLOR_RESET)"
	@echo "  Service Name:  $(COLOR_BLUE)$(SERVICE_NAME)$(COLOR_RESET)"
	@echo "  Binary Name:   $(COLOR_BLUE)$(BIN_NAME)$(COLOR_RESET)"
	@echo "  Binary Path:   $(COLOR_BLUE)$(BIN_DIR)/$(BIN_NAME)$(COLOR_RESET)"
	@echo "  CMD Path:      $(COLOR_BLUE)$(CMD_PATH)$(COLOR_RESET)"
	@echo "  Go Version:    $(COLOR_BLUE)$$(go version | cut -d' ' -f3)$(COLOR_RESET)"

.DEFAULT_GOAL := help
