package cmd

import (
	"log"
	"os"

	"github.com/minehubmc/plugin-downloader/internal"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(pluginCmd)
}

var pluginCmd = &cobra.Command{
	TraverseChildren: true,
	Use:              "plugins",
	Short:            "Download all of the plugins",
	Long:             "It first downloads all of the plugins listed in the root dependencies.json. After that it reads the dependencies.json (if exists) in each of those jars. If one of those plugins hasn't been downloaded, it downloads it. If there are version conflicts an error is reported.",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := zap.NewDevelopment()

		if err != nil {
			log.Fatal("failed to create zap logger", err)
		}

		config := internal.Parse(configFilePath, logger)

		if err := internal.PrepareOutputFolder(outputFolder, logger); err != nil {
			logger.Fatal("Failed to prepare output folder", zap.Error(err))
		}

		filterTags, err := rootCmd.PersistentFlags().GetStringSlice("tags")
		if err != nil {
			logger.Fatal("Failed to parse filter tags", zap.Error(err))
		}

		errs := internal.DownloadPlugins(internal.GetPlugins(config), config.Credentials, outputFolder, logger, filterTags, localM2RepoPath)
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
