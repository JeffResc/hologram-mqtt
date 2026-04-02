// Package main is the entrypoint for the hologram-mqtt bridge.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	hc := hologram.NewClient(cfg.Hologram.APIKey, logger)

	mc, err := mqtt.NewClient(mqtt.ClientConfig{
		Broker:      cfg.MQTT.Broker,
		Username:    cfg.MQTT.Username,
		Password:    cfg.MQTT.Password,
		ClientID:    cfg.MQTT.ClientID,
		TopicPrefix: cfg.MQTT.TopicPrefix,
	}, logger)
	if err != nil {
		logger.Error("failed to connect to MQTT broker", "error", err)
		os.Exit(1)
	}

	b := bridge.New(cfg, hc, mc, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := b.Run(ctx); err != nil {
		logger.Error("bridge error", "error", err)
		os.Exit(1)
	}

	mc.Disconnect()
	logger.Info("hologram-mqtt stopped")
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
}
