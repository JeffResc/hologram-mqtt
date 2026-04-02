package mqtt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientConfigDefaults(t *testing.T) {
	cfg := ClientConfig{
		Broker:      "tcp://localhost:1883",
		ClientID:    "test-client",
		TopicPrefix: "test-prefix",
	}

	assert.Equal(t, "tcp://localhost:1883", cfg.Broker)
	assert.Equal(t, "test-client", cfg.ClientID)
	assert.Equal(t, "test-prefix", cfg.TopicPrefix)
}

func TestMockPublisher(t *testing.T) {
	mock := NewMockPublisher()
	assert.True(t, mock.IsConnected())

	err := mock.Publish("test/topic", 1, true, []byte("hello"))
	assert.NoError(t, err)
	assert.Len(t, mock.Published, 1)
	assert.Equal(t, "test/topic", mock.Published[0].Topic)
	assert.Equal(t, []byte("hello"), mock.Published[0].Payload)

	err = mock.Subscribe("test/sub", 1, func(topic string, payload []byte) {})
	assert.NoError(t, err)
	assert.Len(t, mock.Subscribed, 1)

	mock.Disconnect()
	assert.False(t, mock.IsConnected())
}
