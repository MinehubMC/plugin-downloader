package cmd

import (
	"fmt"
	"os"

	"github.com/minehubmc/plugin-downloader/internal"
	"github.com/spf13/cobra"
)

var (
	configFilePath string
	outputFolder   string

	rootCmd = &cobra.Command{
		Use:   "pld",
		Short: "Automatically download required plugins/dependencies for minecraft servers.",
		Long:  "It reads a .json file and downloads the plugins to a specified folder. Created for easier creation of docker images.",
		Run: func(cmd *cobra.Command, args []string) {
			internal.Parse(configFilePath)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&configFilePath, "config", "dependencies.json", "config file (default is $PWD/dependencies.json)")
	rootCmd.PersistentFlags().StringVar(&outputFolder, "out", ".", "download output folder (default is .)")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
