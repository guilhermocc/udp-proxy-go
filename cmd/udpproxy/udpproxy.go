package udpproxy

import (
	"context"
	"simple-udp-proxy/internal/proxy"
	"simple-udp-proxy/internal/serviceconfig"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	logConfig  string
	configPath string
)

var ProxyCmd = &cobra.Command{
	Use:     "udp-proxy",
	Short:   "Starts an udp proxy",
	Example: "udpproxy start udp-proxy -l production -c config/udp-proxy.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		runWorker()
	},
}

func init() {
	// Define cmd flags here
	ProxyCmd.Flags().StringVarP(&logConfig, "log-config", "l", "production", "preset of configurations used by the logs. possible values are \"development\" or \"production\".")
	ProxyCmd.Flags().StringVarP(&configPath, "config-path", "c", "config/udp-proxy.yaml", "path of the configuration YAML file")
}

func runWorker() {
	ctx, cancelFn := context.WithCancel(context.Background())

	err, _, shutdownInternalServerFn := serviceconfig.ServiceSetup(ctx, cancelFn, logConfig, configPath)
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("unable to setup service")
	}

	zap.L().Info("Starting proxy server...")
	proxy.RunProxy()

	<-ctx.Done()

	err = shutdownInternalServerFn()
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("failed to shutdown internal server")
	}
}
