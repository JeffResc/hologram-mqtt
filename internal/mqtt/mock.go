package mqtt

// MockPublisher implements Publisher for testing other packages.
type MockPublisher struct {
	Published    []PublishCall
	Subscribed   []SubscribeCall
	Connected    bool
	PublishErr   error
	SubscribeErr error
}

// PublishCall records a call to Publish.
type PublishCall struct {
	Topic    string
	QoS      byte
	Retained bool
	Payload  []byte
}

// SubscribeCall records a call to Subscribe.
type SubscribeCall struct {
	Topic   string
	QoS     byte
	Handler MessageHandler
}

// NewMockPublisher creates a new MockPublisher that reports as connected.
func NewMockPublisher() *MockPublisher {
	return &MockPublisher{Connected: true}
}

func (m *MockPublisher) Publish(topic string, qos byte, retained bool, payload []byte) error {
	if m.PublishErr != nil {
		return m.PublishErr
	}
	m.Published = append(m.Published, PublishCall{
		Topic:    topic,
		QoS:      qos,
		Retained: retained,
		Payload:  append([]byte{}, payload...),
	})
	return nil
}

func (m *MockPublisher) Subscribe(topic string, qos byte, handler MessageHandler) error {
	if m.SubscribeErr != nil {
		return m.SubscribeErr
	}
	m.Subscribed = append(m.Subscribed, SubscribeCall{
		Topic:   topic,
		QoS:     qos,
		Handler: handler,
	})
	return nil
}

func (m *MockPublisher) IsConnected() bool {
	return m.Connected
}

func (m *MockPublisher) Disconnect() {
	m.Connected = false
}

// FindPublished returns all publish calls matching the given topic.
func (m *MockPublisher) FindPublished(topic string) []PublishCall {
	var found []PublishCall
	for _, p := range m.Published {
		if p.Topic == topic {
			found = append(found, p)
		}
	}
	return found
}
