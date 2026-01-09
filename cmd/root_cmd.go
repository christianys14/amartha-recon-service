package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "billingService",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
	examples and usage of using your application. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println("error on execute rootCmd: " + err.Error())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(
		serveHttp,
	)
}

func initConfig() {
	// We can move the initialization of the config here instead one by one inside each command
}
