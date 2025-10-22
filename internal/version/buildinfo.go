/*
Copyright © 2025 Erick Wachira (tesh254)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package version

import (
	"runtime/debug"
)

// BuildInfo variables that will be set during build time
var (
	BuildVersion = "dev"      // will be set to git tag during release builds
	BuildCommit  = "unknown"  // will be set to git commit hash during build
	BuildDate    = "unknown"  // will be set to build timestamp during build
)

// GetBuildInfo retrieves build information from buildinfo or fallback variables
func GetBuildInfo() (version, commit, date string) {
	// Try to get version from Go build info first
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				commit = setting.Value
			case "vcs.time":
				date = setting.Value
			}
		}
		
		// Check if this is a tagged version
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		} else {
			// Fall back to build variables
			version = BuildVersion
			if commit == "" {
				commit = BuildCommit
			}
			if date == "" {
				date = BuildDate
			}
		}
	} else {
		// Fall back to build variables if buildinfo is not available
		version = BuildVersion
		commit = BuildCommit
		date = BuildDate
	}
	
	return version, commit, date
}