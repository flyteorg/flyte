package version

import (
	"time"

	"github.com/sirupsen/logrus"
)

// This module provides the ability to inject Build (git sha) and Version information at compile time.
// To set these values invoke go build as follows
// go build -ldflags “-X github.com/lyft/flytepropeller/version.Build=xyz -X github.com/lyft/flytepropeller/version.Version=1.2.3"
// will provide the build and version information
var (
	// Specifies the GIT sha of the build
	Build = "unknown"
	// Version for the build, should follow a semver
	Version = "unknown"
	// Build timestamp
	BuildTime = time.Now().String()
)

// LogBuildInformation Use this method to log the build information for the current app. The app name should be provided. To inject the build
// and version information refer to the top-level comment in this file
func LogBuildInformation(appName string) {
	logrus.Info("------------------------------------------------------------------------")
	logrus.Infof("App [%s], Version [%s], BuildSHA [%s], BuildTS [%s]", appName, Version, Build, BuildTime)
	logrus.Info("------------------------------------------------------------------------")
}
