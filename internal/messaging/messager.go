package messaging

import "context"

type Message interface {
	Topic() string
	Data() (map[string]any, error)
	Commit()
}

type MessageConsumer interface {
	SubscribeInTopic(ctx context.Context, topic string) error
	Consume(ctx context.Context, topic string) (Message, error)
}

type MessageProducer interface {
	Publish(ctx context.Context, topic string, messageData map[string]any) error
}
