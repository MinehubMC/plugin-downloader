package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/minehubmc/plugin-downloader/internal"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	configFilePath  string
	outputFolder    string
	localM2RepoPath string

	rootCmd = &cobra.Command{
		Use:   "pld",
		Short: "Automatically download required plugins/dependencies for minecraft servers.",
		Long:  "It reads a .json file and downloads the plugins to a specified folder. Created for easier creation of docker images.",
		Run: func(cmd *cobra.Command, args []string) {
			logger, err := zap.NewDevelopment()

			if err != nil {
				log.Fatal("failed to create zap logger", err)
			}

			config := internal.Parse(configFilePath, logger)
			if err := internal.PrepareOutputFolder(outputFolder, logger); err != nil {
				log.Fatal(err)
			}

			filterTags, err := cmd.PersistentFlags().GetStringSlice("tags")

			if err != nil {
				log.Fatal(err)
			}

			errs := internal.Download(config, outputFolder, logger, filterTags, localM2RepoPath)

			if len(errs) > 0 {
				for _, err := range errs {
					logger.Error(err.Error())
				}
				os.Exit(1)
			} else {
				os.Exit(0)
			}
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&configFilePath, "config", "dependencies.json", "config file (default is $PWD/dependencies.json)")
	rootCmd.PersistentFlags().StringVar(&outputFolder, "out", ".", "download output folder (default is .)")
	rootCmd.PersistentFlags().StringSlice("tags", []string{}, "comma-separated list of tags to filter plugins by")
	rootCmd.PersistentFlags().StringVar(&localM2RepoPath, "local-maven-repository", "~/.m2", "change the repository where dependencies are installed to")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
