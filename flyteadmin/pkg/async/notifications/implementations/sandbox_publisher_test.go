package implementations

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// mockMessage is a dummy proto message that will always fail to marshal
type mockMessage struct{}

func (m *mockMessage) Reset()                   {}
func (m *mockMessage) String() string           { return "mockMessage" }
func (m *mockMessage) ProtoMessage()            {}
func (m *mockMessage) Marshal() ([]byte, error) { return nil, errors.New("forced marshal error") }

func TestSandboxPublisher_Publish(t *testing.T) {
	msgChan := make(chan []byte, 1)
	publisher := NewSandboxPublisher(msgChan)

	err := publisher.Publish(context.Background(), "NOTIFICATION_TYPE", &testEmail)

	assert.NotZero(t, len(msgChan))
	assert.Nil(t, err)
}

func TestSandboxPublisher_PublishMarshalError(t *testing.T) {
	msgChan := make(chan []byte, 1)
	publisher := NewSandboxPublisher(msgChan)

	err := publisher.Publish(context.Background(), "testMarshallError", &mockMessage{})
	assert.Error(t, err)
	assert.Equal(t, "forced marshal error", err.Error())
}
