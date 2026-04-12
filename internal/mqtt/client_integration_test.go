//go:build integration

package mqtt_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jeffresc/hologram-mqtt/internal/mqtt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startMosquitto(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "eclipse-mosquitto:2",
		ExposedPorts: []string{"1883/tcp"},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            strings.NewReader("listener 1883\nallow_anonymous true\n"),
				ContainerFilePath: "/mosquitto/config/mosquitto.conf",
			},
		},
		WaitingFor: wait.ForListeningPort("1883/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "1883/tcp")
	require.NoError(t, err)

	broker := fmt.Sprintf("tcp://%s:%s", host, port.Port())

	cleanup := func() {
		_ = container.Terminate(ctx)
	}

	return broker, cleanup
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestIntegrationConnectDisconnect(t *testing.T) {
	broker, cleanup := startMosquitto(t)
	defer cleanup()

	client, err := mqtt.NewClient(mqtt.ClientConfig{
		Broker:      broker,
		ClientID:    "test-connect",
		TopicPrefix: "test",
	}, testLogger())
	require.NoError(t, err)

	assert.True(t, client.IsConnected())
	client.Disconnect()
}

func TestIntegrationPublishSubscribeRoundTrip(t *testing.T) {
	broker, cleanup := startMosquitto(t)
	defer cleanup()

	// Create publisher
	pub, err := mqtt.NewClient(mqtt.ClientConfig{
		Broker:      broker,
		ClientID:    "test-pub",
		TopicPrefix: "test",
	}, testLogger())
	require.NoError(t, err)
	defer pub.Disconnect()

	// Create subscriber
	sub, err := mqtt.NewClient(mqtt.ClientConfig{
		Broker:      broker,
		ClientID:    "test-sub",
		TopicPrefix: "test",
	}, testLogger())
	require.NoError(t, err)
	defer sub.Disconnect()

	var mu sync.Mutex
	var received []string

	err = sub.Subscribe("test/device/+/state", 1, func(topic string, payload []byte) {
		mu.Lock()
		received = append(received, string(payload))
		mu.Unlock()
	})
	require.NoError(t, err)

	// Give subscription time to propagate
	time.Sleep(200 * time.Millisecond)

	// Publish messages
	for i := 0; i < 3; i++ {
		err = pub.Publish(fmt.Sprintf("test/device/%d/state", i), 1, false, []byte(fmt.Sprintf("msg-%d", i)))
		require.NoError(t, err)
	}

	// Wait for messages to arrive
	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(received) == 3
	}, 5*time.Second, 100*time.Millisecond, "expected 3 messages, got %d", len(received))
}

func TestIntegrationRetainedMessage(t *testing.T) {
	broker, cleanup := startMosquitto(t)
	defer cleanup()

	// Publish a retained message
	pub, err := mqtt.NewClient(mqtt.ClientConfig{
		Broker:      broker,
		ClientID:    "test-retain-pub",
		TopicPrefix: "test",
	}, testLogger())
	require.NoError(t, err)

	err = pub.Publish("test/status", 1, true, []byte("online"))
	require.NoError(t, err)
	pub.Disconnect()

	// New subscriber should receive the retained message
	sub, err := mqtt.NewClient(mqtt.ClientConfig{
		Broker:      broker,
		ClientID:    "test-retain-sub",
		TopicPrefix: "test",
	}, testLogger())
	require.NoError(t, err)
	defer sub.Disconnect()

	var got string
	var mu sync.Mutex

	err = sub.Subscribe("test/status", 1, func(_ string, payload []byte) {
		mu.Lock()
		got = string(payload)
		mu.Unlock()
	})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return got == "online"
	}, 5*time.Second, 100*time.Millisecond)
}
