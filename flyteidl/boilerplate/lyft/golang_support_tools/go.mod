module github.com/lyft/boilerplate

go 1.13

require (
	github.com/alvaroloes/enumer v1.1.2
	github.com/golangci/golangci-lint v1.22.2
	github.com/lyft/flytestdlib v0.2.31
	github.com/pseudomuto/protoc-gen-doc v1.4.1
	github.com/vektra/mockery v0.0.0-20181123154057-e78b021dcbb5
)

replace github.com/vektra/mockery => github.com/enghabu/mockery v0.0.0-20191009061720-9d0c8670c2f0

replace github.com/pseudomuto/protoc-gen-doc => github.com/paweld2/protoc-gen-doc v1.4.2-0.20210329170919-e6b3293482d4
