# hologram-mqtt

Hologram.io MQTT Bridge for Home Assistant — exposes your Hologram cellular IoT devices as native Home Assistant devices via MQTT Auto Discovery.

## Features

- **Automatic device discovery** — Hologram devices appear automatically in Home Assistant via MQTT Discovery
- **Real-time status** — Device state, connectivity, carrier, plan, IMEI, and data usage published to MQTT
- **Pause/Resume control** — Toggle device data via Home Assistant switches (calls Hologram API)
- **Configurable polling** — Device info refreshed on a configurable interval (default: 5 minutes)
- **TLS support** — Optional TLS with custom CA, client certificates, and skip-verify for MQTT connections
- **Health check** — Built-in `/healthz` HTTP endpoint for container orchestrators
- **Prometheus metrics** — Built-in `/metrics` endpoint for observability
- **Lightweight** — Runs as a single Go binary in a distroless Docker container

## Entities Per Device

| Entity Type | Name | Description |
|---|---|---|
| Sensor | State | Device state (LIVE, PAUSED, DEAD) |
| Sensor | IMEI | Device IMEI number |
| Sensor | SIM Number | SIM card number |
| Sensor | Carrier | Cellular carrier |
| Sensor | Plan | Data plan name |
| Sensor | Phone Number | Device phone number |
| Sensor | Last Connection | Last connection timestamp |
| Sensor | Network | Network technology (LTE, etc.) |
| Sensor | Data Up (bytes) | Bytes uploaded in recent session |
| Sensor | Data Down (bytes) | Bytes downloaded in recent session |
| Binary Sensor | Connectivity | Online when device state is LIVE |
| Switch | Active | Toggle to pause/resume the device |

## Quick Start

### Docker (Recommended)

```bash
docker run -d \
  --name hologram-mqtt \
  -e HOLOGRAM_API_KEY=your_api_key \
  -e MQTT_BROKER=tcp://your-mqtt-broker:1883 \
  ghcr.io/jeffresc/hologram-mqtt:latest
```

### Docker Compose

```yaml
services:
  hologram-mqtt:
    image: ghcr.io/jeffresc/hologram-mqtt:latest
    restart: unless-stopped
    environment:
      HOLOGRAM_API_KEY: "your_api_key"
      MQTT_BROKER: "tcp://mqtt:1883"
      MQTT_USERNAME: ""
      MQTT_PASSWORD: ""
      POLL_INTERVAL: "5m"
```

### Kubernetes (Helm)

The Helm chart is published as an OCI artifact to `ghcr.io`.

```bash
helm install hologram-mqtt oci://ghcr.io/jeffresc/charts/hologram-mqtt \
  --set hologram.apiKey=your_api_key \
  --set mqtt.broker=tcp://mqtt:1883
```

To install a specific version:

```bash
helm install hologram-mqtt oci://ghcr.io/jeffresc/charts/hologram-mqtt --version 1.0.0
```

#### Using an Existing Secret

Instead of passing the API key directly, you can reference a pre-existing Kubernetes secret:

```bash
kubectl create secret generic hologram-api \
  --from-literal=api-key=your_api_key

helm install hologram-mqtt oci://ghcr.io/jeffresc/charts/hologram-mqtt \
  --set hologram.existingSecret=hologram-api \
  --set mqtt.broker=tcp://mqtt:1883
```

Similarly for MQTT credentials:

```bash
kubectl create secret generic mqtt-creds \
  --from-literal=username=myuser \
  --from-literal=password=mypass

helm install hologram-mqtt oci://ghcr.io/jeffresc/charts/hologram-mqtt \
  --set hologram.apiKey=your_api_key \
  --set mqtt.broker=tcp://mqtt:1883 \
  --set mqtt.existingSecret=mqtt-creds
```

#### Custom Values File

For more complex configurations, create a `values.yaml` file:

```yaml
hologram:
  apiKey: "your_api_key"

mqtt:
  broker: "tcp://mqtt:1883"
  username: "user"
  password: "pass"
  tls:
    enabled: true
    caSecret: "mqtt-ca"
    clientCertSecret: "mqtt-client-cert"

pollInterval: "10m"
logLevel: "debug"
```

```bash
helm install hologram-mqtt oci://ghcr.io/jeffresc/charts/hologram-mqtt -f values.yaml
```

See [`chart/values.yaml`](chart/values.yaml) for all available options.

### Binary

```bash
go install github.com/jeffresc/hologram-mqtt/cmd/hologram-mqtt@latest

export HOLOGRAM_API_KEY=your_api_key
export MQTT_BROKER=tcp://localhost:1883
hologram-mqtt
```

## Configuration

Configuration is loaded from a YAML file and/or environment variables. Environment variables take precedence over the config file.

| Environment Variable | Config Key | Default | Description |
|---|---|---|---|
| `HOLOGRAM_API_KEY` | `hologram.api_key` | (required) | Hologram API key |
| `HOLOGRAM_ORG_ID` | `hologram.org_id` | | Organization ID (if multi-org) |
| `MQTT_BROKER` | `mqtt.broker` | (required) | MQTT broker address |
| `MQTT_USERNAME` | `mqtt.username` | | MQTT username |
| `MQTT_PASSWORD` | `mqtt.password` | | MQTT password |
| `MQTT_CLIENT_ID` | `mqtt.client_id` | `hologram-mqtt` | MQTT client ID |
| `MQTT_TOPIC_PREFIX` | `mqtt.topic_prefix` | `hologram` | MQTT topic prefix |
| `MQTT_TLS_ENABLED` | `mqtt.tls.enabled` | `false` | Enable TLS |
| `MQTT_TLS_CA_CERT` | `mqtt.tls.ca_cert` | | Path to CA certificate |
| `MQTT_TLS_CLIENT_CERT` | `mqtt.tls.client_cert` | | Path to client certificate |
| `MQTT_TLS_CLIENT_KEY` | `mqtt.tls.client_key` | | Path to client private key |
| `MQTT_TLS_SKIP_VERIFY` | `mqtt.tls.skip_verify` | `false` | Skip TLS verification |
| `DISCOVERY_PREFIX` | `discovery.prefix` | `homeassistant` | HA discovery prefix |
| `DISCOVERY_ENABLED` | `discovery.enabled` | `true` | Enable HA discovery |
| `HEALTH_ENABLED` | `health.enabled` | `true` | Enable health endpoint |
| `HEALTH_ADDR` | `health.addr` | `:8080` | Health server listen address |
| `POLL_INTERVAL` | `poll_interval` | `5m` | Polling interval |
| `LOG_LEVEL` | `log_level` | `info` | Log level |
| `CONFIG_FILE` | | `config.yaml` | Path to config file |

See [config.example.yaml](config.example.yaml) for a documented example.

## MQTT Topics

```
hologram/status                              → Bridge online/offline (LWT)
hologram/device/<id>/availability            → Device availability
hologram/device/<id>/attributes              → Device attributes (JSON)
hologram/device/<id>/connectivity            → ON/OFF
hologram/device/<id>/switch/state            → ON/OFF (active/paused)
hologram/device/<id>/switch/set              → Command topic (ON/OFF)
```

## Metrics

Prometheus metrics are exposed at `/metrics` on the health HTTP server (same port as `/healthz`, default `:8080`).

| Metric | Type | Description |
|---|---|---|
| `hologram_mqtt_polls_total` | Counter | Total poll cycles (`status=success\|error`) |
| `hologram_mqtt_poll_duration_seconds` | Histogram | Duration of each poll cycle |
| `hologram_mqtt_devices_total` | Gauge | Number of known devices after last poll |
| `hologram_mqtt_commands_total` | Counter | MQTT commands received (`action=live\|pause`) |

To scrape metrics in Kubernetes, create a `ServiceMonitor` or `PodMonitor` CR targeting this service, or add `prometheus.io/*` pod annotations via `podAnnotations` in the Helm values.

## Development

### Prerequisites

- Go 1.26+

### Build

```bash
go build -o hologram-mqtt ./cmd/hologram-mqtt
```

### Test

```bash
go test -race ./...
```

### Lint

```bash
golangci-lint run
```

## License

[Apache 2.0](LICENSE)
