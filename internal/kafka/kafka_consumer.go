package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConsumer struct {
	kafkaConsumer *kafka.Consumer
}

var bootstrapServers = "localhost:9092"

func NewKafkaConsumer() *KafkaConsumer {
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

func (k *KafkaConsumer) Consume(ctx context.Context, topic string) (map[string]any, error) {
	message, err := k.kafkaConsumer.ReadMessage(-1)
	if err != nil {
		panic(err)
	}
	jsonData := map[string]any{}
	err = json.Unmarshal(message.Value, &jsonData)
	if err != nil {
		log.Fatalf("Erro ao deserializar mensagem: %v\n", err)
		return nil, err
	}
	return jsonData, nil
}
