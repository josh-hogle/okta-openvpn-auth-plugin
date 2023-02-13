package app

import (
	"github.com/Masterminds/semver"
)

// General app settings.
const (
	Title        = "Okta OpenVPN Auth Plugin"
	Copyright    = "Copyright (c) 2021 Josh Hogle.  All rights reserved."
	EnvVarPrefix = "OKTA_OPENVPN_AUTH_PLUGIN_"
)

var (
	// Build is the first 8 characters of the git commit hash.
	Build string

	// Config holds the application configuration settings.
	Config *config

	// DevBuild is a flag to indicate if this is a developer build.
	DevBuild bool

	// DevBuildStr is the string version of DevBuild which is passed in at compile-time.
	DevBuildStr string

	// SemanticVersion is the actual semantic version of the product.
	SemanticVersion *semver.Version

	// Version is the current semver-compatible version of the product.
	Version string
)

func init() {
}
