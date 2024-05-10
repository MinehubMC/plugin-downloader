package cmd

import (
	"log"
	"os"

	"github.com/minehubmc/plugin-downloader/internal"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(libraryCmd)
}

var libraryCmd = &cobra.Command{
	TraverseChildren: true,
	Use:              "libraries",
	Short:            "Download the libraries, 0 plugins",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := zap.NewDevelopment()

		if err != nil {
			log.Fatal("failed to create zap logger", err)
		}

		config := internal.Parse(configFilePath, logger)

		if len(config.Libraries) <= 0 {
			logger.Fatal("no libraries defined")
		}

		if err := internal.PrepareOutputFolder(outputFolder, logger); err != nil {
			logger.Fatal("Failed to prepare output folder", zap.Error(err))
		}

		filterTags, err := rootCmd.PersistentFlags().GetStringSlice("tags")
		if err != nil {
			logger.Fatal("Failed to parse filter tags", zap.Error(err))
		}

		errs := internal.Download(config.Libraries, config.Credentials, outputFolder, logger, filterTags, localM2RepoPath)
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
