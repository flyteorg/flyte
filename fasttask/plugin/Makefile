.PHONY: lint
lint:
	@GL_DEBUG=linters_output,env golangci-lint run --deadline=5m -v
	
.PHONY: format
format:
	@GL_DEBUG=linters_output,env golangci-lint run --deadline=5m -v --fix
