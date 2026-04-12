// Package main is the entrypoint for the hologram-mqtt bridge.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "go.uber.org/automaxprocs"

	"github.com/jeffresc/hologram-mqtt/internal/bridge"
	"github.com/jeffresc/hologram-mqtt/internal/config"
	"github.com/jeffresc/hologram-mqtt/internal/discovery"
	"github.com/jeffresc/hologram-mqtt/internal/hologram"
	"github.com/jeffresc/hologram-mqtt/internal/mqtt"
)

// version is set at build time via ldflags.
var version = "dev"

func main() {
	// Set discovery version from build-time variable
	discovery.Version = version

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.LogLevel)
	logger.Info("starting hologram-mqtt", "version", version)

	var holoOpts []hologram.Option
	if cfg.Hologram.OrgID > 0 {
		holoOpts = append(holoOpts, hologram.WithOrgID(cfg.Hologram.OrgID))
	}
	hc := hologram.NewClient(cfg.Hologram.APIKey, logger, holoOpts...)

	mc, err := mqtt.NewClient(mqtt.ClientConfig{
		Broker:      cfg.MQTT.Broker,
		Username:    cfg.MQTT.Username,
		Password:    cfg.MQTT.Password,
		ClientID:    cfg.MQTT.ClientID,
		TopicPrefix: cfg.MQTT.TopicPrefix,
		TLS: mqtt.TLSConfig{
			Enabled:    cfg.MQTT.TLS.Enabled,
			CACert:     cfg.MQTT.TLS.CACert,
			ClientCert: cfg.MQTT.TLS.ClientCert,
			ClientKey:  cfg.MQTT.TLS.ClientKey,
			SkipVerify: cfg.MQTT.TLS.SkipVerify,
		},
	}, logger)
	if err != nil {
		logger.Error("failed to connect to MQTT broker", "error", err)
		os.Exit(1)
	}

	b := bridge.New(cfg, hc, mc, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start health check server
	if cfg.Health.Enabled {
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", b.HealthHandler())
		mux.Handle("/metrics", promhttp.Handler())
		healthServer := &http.Server{Addr: cfg.Health.Addr, Handler: mux}

		go func() {
			logger.Info("starting health check server", "addr", cfg.Health.Addr)
			if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("health check server error", "error", err)
			}
		}()

		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := healthServer.Shutdown(shutdownCtx); err != nil {
				logger.Error("failed to shut down health server", "error", err)
			}
		}()
	}

	if err := b.Run(ctx); err != nil {
		mc.Disconnect()
		logger.Error("bridge error", "error", err)
		return
	}

	mc.Disconnect()
	logger.Info("hologram-mqtt stopped")
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		logLevel = slog.LevelInfo
	}

	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
}
