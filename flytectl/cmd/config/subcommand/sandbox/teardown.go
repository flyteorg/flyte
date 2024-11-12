package sandbox

import (
	"fmt"

	"github.com/spf13/pflag"
)

type TeardownFlags struct {
	Volume bool
}

var (
	DefaultTeardownFlags = &TeardownFlags{}
)

func (f *TeardownFlags) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("TeardownFlags", pflag.ExitOnError)
	cmdFlags.BoolVarP(&f.Volume, fmt.Sprintf("%v%v", prefix, "volume"), "v", f.Volume, "Optional. Clean up Docker volume. This will result in a permanent loss of all data within the database and object store. Use with caution!")
	return cmdFlags
}
