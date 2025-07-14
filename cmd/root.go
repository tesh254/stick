package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tesh254/stick/internal/constants"
	"github.com/tesh254/stick/internal/version"
)

var rootCmd = &cobra.Command{
	Use:     "stick",
	Short:   "stick is a lightweight CLI tool for managing multiple virtual branches in Git, enabling seamless work on different features without branch switching.",
	Version: constants.VERSION(),
	Aliases: []string{"stk"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			fmt.Println(constants.DETAILED_VERSION())
			return nil
		}

		if cmd.Flags().NFlag() == 0 && len(args) == 0 {
			fmt.Print(constants.ASCII)
			fmt.Println(constants.CurrentOSWithVersion())
			fmt.Printf("\n%s\n", constants.GetReleaseInfo())
		}

		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version information",
	Long: `show version information for stick.
This command displays version information extracted automatically from 
the Go build system, including Git commit, build date, and more.`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonFlag, _ := cmd.Flags().GetBool("json")
		shortFlag, _ := cmd.Flags().GetBool("short")
		commitFlag, _ := cmd.Flags().GetBool("commit")

		switch {
		case jsonFlag:
			fmt.Println(version.GetJSONVersion())
		case shortFlag:
			fmt.Println(version.GetShortVersion())
		case commitFlag:
			fmt.Println(version.GetVersionWithCommit())
		default:
			fmt.Println(version.GetDetailedVersion())

			// Add extra info for development builds
			if version.IsDevelopment() {
				fmt.Printf("\n%sNote:%s This is a development build.\n",
					"\033[33m", "\033[0m")
			}
		}
	},
}

// Build info command for detailed build information
var buildInfoCmd = &cobra.Command{
	Use:   "buildinfo",
	Short: "show detailed build information",
	Long:  `show comprehensive build information including module details, VCS info, and build settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		info := version.GetBuildInfo()

		fmt.Printf("Build Information:\n")
		fmt.Printf("==================\n")
		fmt.Printf("Version:      %s\n", info.Version)
		fmt.Printf("Git Commit:   %s\n", info.GitCommit)
		if info.GitTag != "unknown" {
			fmt.Printf("Git Tag:      %s\n", info.GitTag)
		}
		fmt.Printf("Build Date:   %s\n", info.BuildDate)
		fmt.Printf("Go Version:   %s\n", info.GoVersion)
		fmt.Printf("Platform:     %s\n", info.Platform)
		fmt.Printf("Compiler:     %s\n", info.Compiler)
		fmt.Printf("Modified:     %t\n", info.IsModified)
		if info.ModulePath != "" {
			fmt.Printf("Module Path:  %s\n", info.ModulePath)
		}
		if info.ModuleSum != "" {
			fmt.Printf("Module Sum:   %s\n", info.ModuleSum)
		}

		// Show build type
		fmt.Printf("\nBuild Type:   ")
		if version.IsRelease() {
			fmt.Printf("%sRelease%s\n", "\033[32m", "\033[0m")
		} else {
			fmt.Printf("%sDevelopment%s\n", "\033[33m", "\033[0m")
		}
	},
}

func Execute() {
	if err := fang.Execute(context.Background(), rootCmd, fang.WithVersion(constants.VERSION())); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	// Root command flags
	rootCmd.Flags().BoolP("version", "v", false, "Print detailed version information")

	// Version command flags
	versionCmd.Flags().Bool("json", false, "Output version information in JSON format")
	versionCmd.Flags().BoolP("short", "s", false, "Output short version only")
	versionCmd.Flags().BoolP("commit", "c", false, "Output version with commit hash")
	rootCmd.AddCommand(buildInfoCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(createBranchCmd)
}

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting home directory: %v\n", err)
		os.Exit(1)
	}

	configDir := filepath.Join(home, ".stick")
	configName := "config"
	configType := "json"

	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
		os.Exit(1)
	}

	// Configure Viper
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AddConfigPath(configDir)

	// Read configuration file, ignore if it doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err = viper.SafeWriteConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating config file: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
			os.Exit(1)
		}
	}
}
