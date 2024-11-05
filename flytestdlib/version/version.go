package version

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// This module provides the ability to inject Build (git sha) and Version information at compile time.
// To set these values invoke go build as follows
// go build -ldflags “-X github.com/lyft/flytestdlib/version.Build=xyz -X github.com/lyft/flytestdlib/version.Version=1.2.3"
// NOTE: If the version is set and server.StartProfilingServerWithDefaultHandlers are initialized then, `/version`
// will provide the build and version information
var (
	// Specifies the GIT sha of the build
	Build = "unknown"
	// Version for the build, should follow a semver
	Version = "unknown"
	// Build timestamp
	BuildTime = time.Now().String()
	// Git branch that was used to build the binary
	GitBranch = ""
)

// Use this method to log the build information for the current app. The app name should be provided. To inject the build
// and version information refer to the top-level comment in this file
func LogBuildInformation(appName string) {
	logrus.Info("------------------------------------------------------------------------")
	msg := fmt.Sprintf("App [%s], Version [%s], BuildSHA [%s], BuildTS [%s]", appName, Version, Build, BuildTime)
	if GitBranch != "" {
		msg += fmt.Sprintf(", Git Branch [%s]", GitBranch)
	}
	logrus.Info(msg)
	logrus.Info("------------------------------------------------------------------------")
}
