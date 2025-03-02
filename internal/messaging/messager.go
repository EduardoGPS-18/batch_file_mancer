package messaging

import "context"

type MessageConsumer interface {
	SubscribeInTopic(ctx context.Context, topic string) error
	Consume(ctx context.Context, topic string) (map[string]any, error)
}

type MessageProducer interface {
	Publish(ctx context.Context, topic string, message map[string]any) error
}
