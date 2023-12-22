package cmd

import (
	"github.com/kuan525/tiger/state"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stateCmd)
}

var stateCmd = &cobra.Command{
	Use: "state",
	Run: StateHandle,
}

func StateHandle(cmd *cobra.Command, args []string) {
	state.RunMain(ConfigPath)
}
