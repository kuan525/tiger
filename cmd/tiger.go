package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var ConfigPath string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(
		&ConfigPath,
		"config",
		"./tiger.yaml",
		"config file (default is ./tiger.yaml)")
}

var rootCmd = &cobra.Command{
	Use:   "tiger",
	Short: "顶级IM系统",
	Run:   tiger,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func tiger(cmd *cobra.Command, args []string) {

}

func initConfig() {

}
