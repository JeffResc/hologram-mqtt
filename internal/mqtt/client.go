// Package mqtt provides an MQTT client wrapper for the hologram-mqtt bridge.
package mqtt

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

// Publisher abstracts MQTT publish/subscribe operations.
type Publisher interface {
	Publish(topic string, qos byte, retained bool, payload []byte) error
	Subscribe(topic string, qos byte, handler MessageHandler) error
	IsConnected() bool
	Disconnect()
}

// MessageHandler is called when a message is received on a subscribed topic.
type MessageHandler func(topic string, payload []byte)

// ClientConfig holds MQTT connection parameters.
type ClientConfig struct {
	Broker      string
	Username    string
	Password    string
	ClientID    string
	TopicPrefix string
}

type client struct {
	paho         pahomqtt.Client
	topicPrefix  string
	logger       *slog.Logger
	mu           sync.Mutex
	subscriptions map[string]pahomqtt.MessageHandler
}

// NewClient creates and connects a new MQTT client with LWT configured.
func NewClient(cfg ClientConfig, logger *slog.Logger) (Publisher, error) {
	c := &client{
		topicPrefix:  cfg.TopicPrefix,
		logger:       logger,
		subscriptions: make(map[string]pahomqtt.MessageHandler),
	}

	opts := pahomqtt.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(cfg.ClientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetMaxReconnectInterval(2 * time.Minute).
		SetCleanSession(false).
		SetOrderMatters(false).
		SetWill(cfg.TopicPrefix+"/status", "offline", 1, true).
		SetOnConnectHandler(c.onConnect).
		SetConnectionLostHandler(c.onConnectionLost)

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	c.paho = pahomqtt.NewClient(opts)

	token := c.paho.Connect()
	if !token.WaitTimeout(30 * time.Second) {
		return nil, fmt.Errorf("MQTT connection timed out")
	}
	if token.Error() != nil {
		return nil, fmt.Errorf("MQTT connection failed: %w", token.Error())
	}

	return c, nil
}

func (c *client) onConnect(_ pahomqtt.Client) {
	c.logger.Info("connected to MQTT broker")

	// Publish birth message
	token := c.paho.Publish(c.topicPrefix+"/status", 1, true, "online")
	token.Wait()

	// Re-subscribe to all topics
	c.mu.Lock()
	subs := make(map[string]pahomqtt.MessageHandler, len(c.subscriptions))
	for topic, handler := range c.subscriptions {
		subs[topic] = handler
	}
	c.mu.Unlock()

	for topic, handler := range subs {
		token := c.paho.Subscribe(topic, 1, handler)
		token.Wait()
		if token.Error() != nil {
			c.logger.Error("failed to re-subscribe", "topic", topic, "error", token.Error())
		}
	}
}

func (c *client) onConnectionLost(_ pahomqtt.Client, err error) {
	c.logger.Warn("MQTT connection lost", "error", err)
}

// Publish sends a message to the given MQTT topic.
func (c *client) Publish(topic string, qos byte, retained bool, payload []byte) error {
	token := c.paho.Publish(topic, qos, retained, payload)
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("publish to %s timed out", topic)
	}
	return token.Error()
}

// Subscribe registers a handler for messages on the given topic.
func (c *client) Subscribe(topic string, qos byte, handler MessageHandler) error {
	pahoHandler := func(_ pahomqtt.Client, msg pahomqtt.Message) {
		handler(msg.Topic(), msg.Payload())
	}

	c.mu.Lock()
	c.subscriptions[topic] = pahoHandler
	c.mu.Unlock()

	token := c.paho.Subscribe(topic, qos, pahoHandler)
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("subscribe to %s timed out", topic)
	}
	return token.Error()
}

// IsConnected returns whether the client is connected to the broker.
func (c *client) IsConnected() bool {
	return c.paho.IsConnectionOpen()
}

// Disconnect publishes the offline status and disconnects from the broker.
func (c *client) Disconnect() {
	token := c.paho.Publish(c.topicPrefix+"/status", 1, true, "offline")
	token.WaitTimeout(5 * time.Second)
	c.paho.Disconnect(250)
	c.logger.Info("disconnected from MQTT broker")
}
