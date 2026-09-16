# Pflags generator

Generate strongly typed CLI flags for the fields of a Go struct, including nested
structs. With the config package, these flags can override configuration values.

## Build and install

From the repository root, build the binary:

```sh
make -C flytestdlib compile
```

The binary is written to `flytestdlib/bin/pflags`. Add that directory to `PATH`
before running `go generate`.

Alternatively, install it into `GOBIN` (or `GOPATH/bin`):

```sh
go install ./flytestdlib/cli/pflags
```

## Usage

Run the generator in the destination package directory:

```sh
pflags MyStruct --package myproject/mypackage
```

The package defaults to the current directory. You can also add a directive to
the file declaring the struct:

```go
//go:generate pflags MyStruct
```

The target must be a struct. Supported field types include basic types,
JSON-unmarshalable types, nested structs, and supported slices and maps.

The generator writes `mystruct_flags.go` and `mystruct_flags_test.go`. By default,
it looks for a variable named `defaultConfig` to provide defaults. Use
`--default-var` to select another variable, or `--bind-default-var` to bind flags
to that variable's fields.

## Field tags

Use the `pflag` tag to specify a default value and help text as a comma-separated
pair. A matching default variable takes precedence over the tag's default value.

```go
type MyStruct struct {
    StringValue string `json:"str" pflag:"\"hello world\",\"life is short\""`
    BoolValue   bool   `json:"bl" pflag:",This is a bool value that will default to false."`
}
```

Use `pflag:"-"` to exclude a field.
