package kafka

import (
	"context"
	"log"
	"os"
	"performatic-file-processor/internal/messaging"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConsumer struct {
	kafkaConsumer *kafka.Consumer
}

func NewKafkaConsumer() *KafkaConsumer {
	var bootstrapServers = os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	kafkaConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  bootstrapServers,
		"group.id":           "file-processor-group",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	})
	if err != nil {
		log.Fatalf("Erro ao criar consumidor: %v\n", err)
		panic(err)
	}
	return &KafkaConsumer{
		kafkaConsumer: kafkaConsumer,
	}
}

func (k *KafkaConsumer) SubscribeInTopic(ctx context.Context, topic string) error {
	var bootstrapServers = os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	err := k.kafkaConsumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatalf("Erro ao se inscrever no tópico: %v\n", err)
	}

	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	})
	if err != nil {
		log.Fatalf("Erro ao criar consumidor: %v\n", err)
	}
	metadata, err := adminClient.GetMetadata(&topic, true, 5000)
	if err != nil {
		log.Fatalf("Erro ao obter metadados: %v\n", err)
		return err
	}
	if _, exists := metadata.Topics[topic]; !exists {
		adminClient.CreateTopics(ctx, []kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}})
	}
	return nil
}

func (k *KafkaConsumer) Consume(ctx context.Context, topic string) (messaging.Message, error) {
	message, err := k.kafkaConsumer.ReadMessage(1)
	if err != nil {
		if err.(kafka.Error).Code() == kafka.ErrTimedOut {
			return nil, err
		}
		panic(err)
	}

	msg := NewKafkaMessage(message, k.kafkaConsumer)

	return msg, nil
}
