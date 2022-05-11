//go:build console
// +build console

package single

import "embed"

//go:embed dist/*
var console embed.FS
