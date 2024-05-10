package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	configFilePath  string
	outputFolder    string
	localM2RepoPath string

	rootCmd = &cobra.Command{
		TraverseChildren: true,
		Use:              "pld",
		Short:            "Automatically download required plugins/dependencies for minecraft servers.",
		Long:             "It reads a .json file and downloads the plugins to a specified folder. Created for easier creation of docker images.",
		Run: func(cmd *cobra.Command, args []string) {
			logger, err := zap.NewDevelopment()

			if err != nil {
				log.Fatal("failed to create zap logger", err)
			}

			logger.Info("Please use subcommands (libraries, plugins) to download the correct stuff.")
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&configFilePath, "config", "dependencies.json", "config file (default is $PWD/dependencies.json)")
	rootCmd.PersistentFlags().StringVar(&outputFolder, "out", ".", "download output folder (default is .)")
	rootCmd.PersistentFlags().StringSlice("tags", []string{}, "comma-separated list of tags to filter items by")
	rootCmd.PersistentFlags().StringVar(&localM2RepoPath, "local-maven-repository", "~/.m2", "change the repository where dependencies are installed to")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
