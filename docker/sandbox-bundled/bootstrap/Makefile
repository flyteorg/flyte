.PHONY: fmt
fmt:
	go mod tidy
	gofmt -w -s .
	golines -w .

.PHONY: check-fmt
check-fmt:
	@[ -z "$(shell gofmt -l .)" ] || ( echo "Not formatted:" && gofmt -l . && exit 1 )

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test -v ./...
