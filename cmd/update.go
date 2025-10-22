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
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
	"github.com/tesh254/stick/internal/version"
)

const (
	githubAPIURL = "https://api.github.com/repos/tesh254/stick/releases/latest"
	userAgent    = "stick-self-update"
)

// GitHubRelease represents the structure of a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update stick to the latest version",
	Long:  `Check for and install the latest version of stick from GitHub releases.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking for updates...")

		// Get current version
		currentVersion, _, _ := version.GetBuildInfo()
		fmt.Printf("Current version: %s\n", currentVersion)

		// Get latest release
		latestRelease, err := getLatestRelease()
		if err != nil {
			fmt.Printf("Error checking for updates: %v\n", err)
			return
		}

		// Compare versions
		shouldUpdate, err := shouldUpdate(currentVersion, latestRelease.TagName)
		if err != nil {
			fmt.Printf("Error comparing versions: %v\n", err)
			return
		}

		if !shouldUpdate {
			fmt.Println("You're already running the latest version!")
			return
		}

		fmt.Printf("New version available: %s\n", latestRelease.TagName)
		fmt.Print("Do you want to update? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Update cancelled.")
			return
		}

		// Find the appropriate asset for the current platform
		assetName := getAssetName(runtime.GOOS, runtime.GOARCH)
		asset := findAsset(latestRelease.Assets, assetName)
		if asset == nil {
			fmt.Printf("No suitable binary found for your platform (%s/%s)\n", runtime.GOOS, runtime.GOARCH)
			return
		}

		fmt.Printf("Downloading %s...\n", asset.Name)

		// Download the asset
		tempFile, err := downloadAsset(asset.DownloadURL)
		if err != nil {
			fmt.Printf("Error downloading asset: %v\n", err)
			return
		}
		defer os.Remove(tempFile)

		// Update the binary
		if err := updateBinary(tempFile); err != nil {
			fmt.Printf("Error updating binary: %v\n", err)
			return
		}

		fmt.Printf("Successfully updated to version %s!\n", latestRelease.TagName)
	},
}

// getLatestRelease fetches the latest release from GitHub API
func getLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(context.Background(), "GET", githubAPIURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// shouldUpdate compares current and latest versions to determine if an update is needed
func shouldUpdate(current, latest string) (bool, error) {
	// Remove 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// Handle development version
	if current == "dev" {
		return true, nil
	}

	// Parse versions
	currentVer, err := semver.NewVersion(current)
	if err != nil {
		return false, fmt.Errorf("error parsing current version: %v", err)
	}

	latestVer, err := semver.NewVersion(latest)
	if err != nil {
		return false, fmt.Errorf("error parsing latest version: %v", err)
	}

	// Return true if latest is greater than current
	return latestVer.GreaterThan(currentVer), nil
}

// getAssetName returns the expected asset name based on OS and architecture
func getAssetName(goos, goarch string) string {
	// Map GOOS and GOARCH to our release naming convention
	osName := map[string]string{
		"darwin":  "darwin",
		"linux":   "linux",
		"windows": "windows",
	}[goos]

	archName := map[string]string{
		"amd64": "amd64",
		"arm64": "arm64",
		"386":   "386",
	}[goarch]

	if osName == "" || archName == "" {
		return ""
	}

	return fmt.Sprintf("stick-%s-%s", osName, archName)
}

// findAsset finds the appropriate asset in the release
func findAsset(assets []struct {
	Name        string "json:\"name\""
	DownloadURL string "json:\"browser_download_url\""
}, assetName string) *struct {
	Name        string "json:\"name\""
	DownloadURL string "json:\"browser_download_url\""
} {
	// First, try exact match
	for _, asset := range assets {
		if asset.Name == assetName || strings.HasPrefix(asset.Name, fmt.Sprintf("%s.", assetName)) {
			return &asset
		}
	}

	// Then, try to match with common archive extensions
	for _, asset := range assets {
		if strings.Contains(asset.Name, assetName) {
			return &asset
		}
	}

	return nil
}

// downloadAsset downloads the specified asset to a temporary file
func downloadAsset(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "stick-update-*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Copy the response body to the temporary file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

// updateBinary replaces the current binary with the downloaded one
func updateBinary(downloadedPath string) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Handle the case where the downloaded file might be an archive (tar.gz)
	updatedPath, err := extractIfArchive(downloadedPath)
	if err != nil {
		return err
	}

	// Backup current binary
	backupPath := execPath + ".backup"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %v", err)
	}

	// Move the new binary to the current location
	if err := os.Rename(updatedPath, execPath); err != nil {
		// Restore backup
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to update binary: %v", err)
	}

	// Make sure the new binary is executable
	if err := os.Chmod(execPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %v", err)
	}

	// Remove backup
	os.Remove(backupPath)

	return nil
}

// extractIfArchive extracts a tar.gz file if the download is an archive
func extractIfArchive(archivePath string) (string, error) {
	// Check if the file is a tar.gz archive
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read the first few bytes to check for gzip header
	header := make([]byte, 10)
	_, err = file.Read(header)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Check if it's a gzip file (first two bytes are 0x1f, 0x8b)
	if len(header) >= 2 && header[0] == 0x1f && header[1] == 0x8b {
		// Reset file pointer
		file.Seek(0, 0)

		// Create gzip reader
		gzr, err := gzip.NewReader(file)
		if err != nil {
			return "", err
		}
		defer gzr.Close()

		// Create tar reader
		tr := tar.NewReader(gzr)

		// Extract the first executable file
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return "", err
			}

			// Check if it's a file and has the expected name pattern
			if header.Typeflag == tar.TypeReg && (strings.HasPrefix(header.Name, "stick-") || header.Name == "stick") {
				// Create a temporary file for the extracted binary
				tempFile, err := os.CreateTemp("", "stick-extracted-*")
				if err != nil {
					return "", err
				}
				defer tempFile.Close()

				// Copy the file content
				_, err = io.Copy(tempFile, tr)
				if err != nil {
					return "", err
				}

				// Make it executable
				if err := os.Chmod(tempFile.Name(), 0755); err != nil {
					return "", err
				}

				return tempFile.Name(), nil
			}
		}
	}

	// If it's not an archive, return the original path
	return archivePath, nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
