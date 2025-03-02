package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
)

type KafkaProducerImpl struct {
	kafkaProducer *kafka.Producer
}

func NewKafkaProducer() *KafkaProducerImpl {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092"})
	if err != nil {
		log.Fatalf("Erro ao criar produtor: %v\n", err)
	}

	return &KafkaProducerImpl{
		kafkaProducer: p,
	}
}

func (k *KafkaProducerImpl) Publish(ctx context.Context, topic string, message map[string]any) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return errors.New("erro ao serializar mensagem")
	}

	deliveryChan := make(chan kafka.Event)
	err = k.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(uuid.New().String()),
		Value: []byte(messageBytes),
	}, deliveryChan)

	if err != nil {
		return err
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)
	close(deliveryChan)

	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}
	return nil
}
