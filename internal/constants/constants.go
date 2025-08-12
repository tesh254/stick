// Package constants provides shared constants for the application.
package constants

import (
	"fmt"
	"runtime"

	"github.com/tesh254/stick/internal/version"
)

// STICK_ASCII is a constant holding the ASCII art for the application.
const STICK_ASCII = `

 ███████╗████████╗██╗ ██████╗██╗  ██╗
 ██╔════╝╚══██╔══╝██║██╔════╝██║ ██╔╝
 ███████╗   ██║   ██║██║     █████╔╝ 
 ╚════██║   ██║   ██║██║     ██╔═██╗ 
 ███████║   ██║   ██║╚██████╗██║  ██╗
 ╚══════╝   ╚═╝   ╚═╝ ╚═════╝╚═╝  ╚═╝
         
`

// VERSION returns the current version from build info.
// It is a simple version string.
func VERSION() string {
	return version.GetVersion()
}

// VERSION_WITH_COMMIT returns version with commit information.
func VERSION_WITH_COMMIT() string {
	return version.GetVersionWithCommit()
}

// SHORT_VERSION returns a concise version string.
func SHORT_VERSION() string {
	return version.GetShortVersion()
}

// DETAILED_VERSION returns comprehensive version information.
func DETAILED_VERSION() string {
	return version.GetDetailedVersion()
}

// BUILD_SUMMARY returns a one-line build summary.
func BUILD_SUMMARY() string {
	return version.GetBuildSummary()
}

// CurrentOSWithVersion returns OS info with version.
func CurrentOSWithVersion() string {
	operatingSystem := runtime.GOOS
	versionInfo := VERSION_WITH_COMMIT()

	buildType := "release"
	if version.IsDevelopment() {
		buildType = "development"
	}

	return fmt.Sprintf(`stick %s (%s build) running on %s`,
		versionInfo, buildType, operatingSystem)
}

// GetReleaseInfo returns release-specific information.
// It indicates whether the build is a release or development build.
func GetReleaseInfo() string {
	if version.IsRelease() {
		return fmt.Sprintf("Release build %s", VERSION())
	}
	return fmt.Sprintf("Development build %s", BUILD_SUMMARY())
}
