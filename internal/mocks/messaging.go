package mocks

import (
	"context"
	"performatic-file-processor/internal/messaging"

	"github.com/stretchr/testify/mock"
)

type KafkaMessageMock struct {
	mock.Mock
}

func NewKafkaMessageMock() *KafkaMessageMock {
	return &KafkaMessageMock{}
}

func (m *KafkaMessageMock) Topic() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *KafkaMessageMock) Data() (map[string]any, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *KafkaMessageMock) Commit() {
	m.Called()
}

type MessageProducerMock struct {
	mock.Mock
}

func (m *MessageProducerMock) Publish(ctx context.Context, topic string, messageData map[string]any) error {
	args := m.Called(ctx, topic, messageData)

	return args.Error(0)
}

type MessageMock struct {
	mock.Mock
}

func NewMessageMock() *MessageMock {
	return &MessageMock{}
}

func (m *MessageMock) Topic() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MessageMock) Data() (map[string]any, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), nil
}

func (m *MessageMock) Commit() {
	m.Called()
}

type MessageConsumerMock struct {
	mock.Mock
}

func NewMessageConsumerMock() *MessageConsumerMock {
	return &MessageConsumerMock{}
}

func (m *MessageConsumerMock) SubscribeInTopic(ctx context.Context, topic string) error {
	args := m.Called(ctx, topic)
	return args.Error(0)
}

func (m *MessageConsumerMock) Consume(ctx context.Context, topic string) (messaging.Message, error) {
	args := m.Called(ctx, topic)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(messaging.Message), args.Error(1)
}
