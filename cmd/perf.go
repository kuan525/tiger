package cmd

import (
	"github.com/kuan525/tiger/perf"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(perfCmd)
	perfCmd.PersistentFlags().Int32Var(&perf.TcpConnNum, "tcp_conn_num", 100000, "tcp 连接的数量，默认10000")
}

var perfCmd = &cobra.Command{
	Use: "perf",
	Run: PerfHandle,
}

func PerfHandle(cmd *cobra.Command, args []string) {
	perf.RunMain()
}
