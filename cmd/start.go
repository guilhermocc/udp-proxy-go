package cmd

import (
	"simple-udp-proxy/cmd/udpproxy"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the provided service",
}

func init() {
	startCmd.AddCommand(udpproxy.ProxyCmd)
}
