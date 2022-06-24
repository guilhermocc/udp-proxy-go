// MIT License
//
// Copyright (c) 2021 TFG Co
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package serviceconfig

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"

	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func ServiceSetup(ctx context.Context, cancelFn context.CancelFunc, logConfig, configPath string) (error, Config, func() error) {
	err := configureLogging(logConfig)
	if err != nil {
		return fmt.Errorf("unable to configure logging: %w", err), nil, nil
	}

	viperConfig, err := newViperConfig(configPath)
	if err != nil {
		return fmt.Errorf("unable to load config: %w", err), nil, nil
	}

	launchTerminatingListenerGoroutine(cancelFn)

	shutdownInternalServerFn := runInternalServer(ctx, viperConfig)

	return nil, viperConfig, shutdownInternalServerFn
}

func runInternalServer(ctx context.Context, configs Config) func() error {
	mux := http.NewServeMux()
	if configs.GetBool("internalApi.healthcheck.enabled") {
		zap.L().Info("adding healthcheck handler to internal API")
		mux.HandleFunc("/health", handleHealth)
		mux.HandleFunc("/healthz", handleHealth)
	}
	if configs.GetBool("internalApi.metrics.enabled") {
		zap.L().Info("adding metrics handler to internal API")
		mux.Handle("/metrics", promhttp.Handler())
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", configs.GetString("internalApi.port")),
		Handler: mux,
	}

	go func() {
		zap.L().Info(fmt.Sprintf("started HTTP internal at :%s", configs.GetString("internalApi.port")))
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			zap.L().With(zap.Error(err)).Fatal("failed to start HTTP internal server")
		}
	}()

	return func() error {
		shutdownCtx, cancelShutdownFn := context.WithTimeout(context.Background(), configs.GetDuration("internalApi.gracefulShutdownTimeout"))
		defer cancelShutdownFn()

		zap.L().Info("stopping HTTP internal server")
		return httpServer.Shutdown(shutdownCtx)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func launchTerminatingListenerGoroutine(cancelFunc context.CancelFunc) {
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		<-sigs
		zap.L().Info("received termination")

		cancelFunc()
	}()
}

func configureLogging(configPreset string) error {
	var cfg zap.Config
	switch configPreset {
	case "development":
		cfg = zap.NewDevelopmentConfig()
	case "production":
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	default:
		return fmt.Errorf("unexpected log_config: %v", configPreset)
	}

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)
	return nil
}

func newViperConfig(configPath string) (Config, error) {
	config := viper.New()
	config.SetEnvPrefix("maestro")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()

	config.SetConfigType("yaml")
	config.SetConfigFile(configPath)
	err := config.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return config, nil
}

// Config is used to fetch configurations using paths. The interface provides
// a way to fetch configurations in specific types. Paths are set in strings and
// can have separated "scopes" using ".". An example of path would be:
// "api.metrics.enabled".
type Config interface {
	// GetString returns the configuration path as a string. Default: ""
	GetString(string) string
	// GetInt returns the configuration path as an int. Default: 0
	GetInt(string) int
	// GetFloat64 returns the configuration path as a float64. Default: 0.0
	GetFloat64(string) float64
	// GetBool returns the configuration path as a boolean. Default: false
	GetBool(string) bool
	// GetDuration returns a time.Duration of the config. Default: 0
	GetDuration(string) time.Duration
}
