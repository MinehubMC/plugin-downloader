package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	configFilePath string

	rootCmd = &cobra.Command{
		Use:   "pld",
		Short: "Automatically download required plugins/dependencies for minecraft servers.",
		Long:  "It reads a .json file and downloads the plugins to a specified folder. Created for easier creation of docker images.",
		Run: func(cmd *cobra.Command, args []string) {
			// do something
			log.Default().Println("hello", args, configFilePath)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&configFilePath, "config", "", "config file (default is $PWD/dependencies.json)")

	cobra.OnInitialize(onInit)
}

func onInit() {
	if configFilePath == "" {
		configFilePath = "dependencies.json"
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
