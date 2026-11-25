package loadtarget

import "github.com/flyteorg/flyte/v2/flytestdlib/cli/pflags/api/testdata/loaddependency"

type Config struct {
	Remote loaddependency.RemoteAlias `json:"remote"`
}

var defaultConfig = Config{
	Remote: loaddependency.Remote{Name: "default-remote"},
}
