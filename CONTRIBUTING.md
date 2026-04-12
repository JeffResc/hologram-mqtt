# Contributing

Thanks for your interest in contributing to hologram-mqtt!

## Dev Setup

**Requirements:** Go 1.26+ and Docker (for integration tests).

```bash
# Build
go build ./cmd/hologram-mqtt

# Run unit tests
go test -race ./...

# Run integration tests (requires Docker)
go test -tags integration -race ./...

# Lint
golangci-lint run
```

## Branch & PR Conventions

- Branch from `main`
- Use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages (e.g., `feat:`, `fix:`, `docs:`)
- Keep PRs focused — one issue per PR when possible
- Include a brief description of the change and link the related issue

## Architecture Overview

```
internal/
├── bridge/      # Orchestrates polling loop and MQTT command handling
├── config/      # Config file (YAML) + env var loading and validation
├── discovery/   # Home Assistant MQTT discovery message publishing
├── hologram/    # Hologram REST API client
└── mqtt/        # MQTT client wrapper (connect, publish, subscribe)
```

- **Config:** Loaded from `config.yaml` (or `CONFIG_FILE` env var), with environment variable overrides. See `config.example.yaml` for all options.
- **Bridge:** The main loop — polls the Hologram API on an interval, publishes device state to MQTT, and subscribes to command topics for controlling devices.
- **Discovery:** Publishes Home Assistant-compatible MQTT discovery configs so devices appear automatically.
- **Hologram:** HTTP client for the Hologram REST API with pagination and retry logic.
- **MQTT:** Thin wrapper around the Paho MQTT client with TLS support and automatic reconnection.

## Testing

- Unit tests are required for new code
- Use the mock in `internal/mqtt/mock.go` for testing MQTT interactions
- Integration tests use the `//go:build integration` build tag and `testcontainers-go` for a real Mosquitto broker
- Run tests with `-race` to catch data races
