package config

import appconfig "github.com/flyteorg/flyte/v2/app/config"

// InternalAppConfig is an alias of the public config type so internal packages
// can import it without depending on the public app/config path directly.
type InternalAppConfig = appconfig.InternalAppConfig
