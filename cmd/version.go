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
package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/tesh254/stick/internal/version"
)

// VersionInfo holds the version information for the application
type VersionInfo struct {
	Version    string `json:"version"`
	Commit     string `json:"commit"`
	BuildDate  string `json:"build_date"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	IsModified bool   `json:"is_modified"`
}

var (
	versionShort bool
	versionJSON  bool
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version, commit hash, build date and other build information.`,
	Run: func(cmd *cobra.Command, args []string) {
		version, commit, date := version.GetBuildInfo()

		// Determine if the build is modified
		isModified := commit == "unknown" || (version == "dev" && commit != "unknown")

		info := VersionInfo{
			Version:    version,
			Commit:     commit,
			BuildDate:  date,
			OS:         runtime.GOOS,
			Arch:       runtime.GOARCH,
			IsModified: isModified,
		}

		if versionJSON {
			// Output JSON format
			jsonData, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				fmt.Printf("Error formatting JSON: %v\n", err)
				return
			}
			fmt.Println(string(jsonData))
		} else if versionShort {
			// Output short format - just the version
			fmt.Println(info.Version)
		} else {
			// Output detailed format
			fmt.Printf("stick version %s\n", info.Version)
			fmt.Printf("commit: %s\n", info.Commit)
			fmt.Printf("build date: %s\n", info.BuildDate)
			fmt.Printf("os: %s\n", info.OS)
			fmt.Printf("arch: %s\n", info.Arch)
			if info.IsModified {
				fmt.Printf("status: modified (development build)\n")
			} else {
				fmt.Printf("status: release\n")
			}
		}
	},
}

// buildinfoCmd represents the buildinfo command
var buildinfoCmd = &cobra.Command{
	Use:   "buildinfo",
	Short: "Print comprehensive build information",
	Long:  `Print detailed build information including version, commit, build date, and platform details.`,
	Run: func(cmd *cobra.Command, args []string) {
		version, commit, date := version.GetBuildInfo()

		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Build Date: %s\n", date)
		fmt.Printf("OS: %s\n", runtime.GOOS)
		fmt.Printf("Architecture: %s\n", runtime.GOARCH)
	},
}

func init() {
	versionCmd.Flags().BoolVar(&versionShort, "short", false, "Print only the version number")
	versionCmd.Flags().BoolVar(&versionJSON, "json", false, "Print version information in JSON format")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(buildinfoCmd)
}
