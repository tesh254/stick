package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tesh254/stick/internal/config"
	"github.com/tesh254/stick/internal/constants"
	"github.com/tesh254/stick/internal/version"
)

// StickConfig holds configuration for the application.
type StickConfig = config.StickConfig

// NewStickConfig is a constructor for StickConfig that reads from Viper.
var NewStickConfig = config.NewStickConfig

var stickCmd = &cobra.Command{
	Use:     "stick",
	Short:   "Stick basically upgrades your git",
	Aliases: []string{"stk"},
	// Use the container's NewRun method to inject dependencies.
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle version flag specially to show detailed info
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			fmt.Println(constants.DETAILED_VERSION())
			return nil
		}

		if cmd.Flags().NFlag() == 0 && len(args) == 0 {
			fmt.Print(constants.STICK_ASCII)
			fmt.Println(constants.CurrentOSWithVersion())
			fmt.Printf("\n%s\n", constants.GetReleaseInfo())
		}
		return nil
	},
}

// Version command with multiple output formats
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
	Short: "Show detailed build information",
	Long:  `Show comprehensive build information including module details, VCS info, and build settings.`,
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

// Execute runs the root command.
func Execute() {
	if err := fang.Execute(context.Background(), stickCmd, fang.WithVersion(constants.VERSION())); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Here you can define flags and bind them to viper.
	stickCmd.PersistentFlags().String("name", "", "a name from a flag")
	stickCmd.PersistentFlags().String("selected_model", "", "selected model e.g. gemini-pro")
	stickCmd.PersistentFlags().String("api_key", "", "api key for the selected model")
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	viper.AddConfigPath(home)
	viper.SetConfigName(".stick")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only print error if it's not a "file not found" error.
			fmt.Printf("Error reading config file: %s", err)
		}
	}
	viper.BindPFlag("name", stickCmd.PersistentFlags().Lookup("name"))
	viper.BindPFlag("selected_model", stickCmd.PersistentFlags().Lookup("selected_model"))
	viper.BindPFlag("api_key", stickCmd.PersistentFlags().Lookup("api_key"))
	cobra.OnInitialize()
	stickCmd.CompletionOptions.DisableDefaultCmd = false

	// Root command flags
	stickCmd.Flags().BoolP("version", "v", false, "Print detailed version information")

	// Version command flags
	versionCmd.Flags().Bool("json", false, "Output version information in JSON format")
	versionCmd.Flags().BoolP("short", "s", false, "Output short version only")
	versionCmd.Flags().BoolP("commit", "c", false, "Output version with commit hash")

	// show command flags
	showCmd.Flags().StringP("repo-path", "r", "current", "directory to check commit diff, --repo-path (optional) auto picks current pwd")
	showCmd.Flags().StringP("commit", "c", "", "commit hash to check diff")

	// release command flags
	releaseCmd.Flags().StringP("repo-path", "r", "current", "directory to push and release")
	releaseCmd.Flags().StringP("commit", "c", "HEAD", "commit hash or default to HEAD")
	releaseCmd.Flags().StringP("message", "m", "", "commit message")
	releaseCmd.Flags().StringP("tag", "t", "", "commit hash to check diff")

	// status command flags
	statusCmd.Flags().StringP("repo-path", "r", "current", "directory to check worktree changes")

	// agent commands
	agentCmd.AddCommand(agentInitCmd)

	stickCmd.AddCommand(buildInfoCmd)
	stickCmd.AddCommand(versionCmd)
	stickCmd.AddCommand(showCmd)
	stickCmd.AddCommand(releaseCmd)
	stickCmd.AddCommand(statusCmd)
	stickCmd.AddCommand(agentCmd)
}
