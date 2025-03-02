package kafka

import (
	"encoding/json"

	kafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaMessage struct {
	message       *kafka.Message
	kafkaConsumer *kafka.Consumer
}

func NewKafkaMessage(message *kafka.Message, consumer *kafka.Consumer) *KafkaMessage {
	return &KafkaMessage{
		message:       message,
		kafkaConsumer: consumer,
	}
}

func (m *KafkaMessage) Topic() string {
	return *m.message.TopicPartition.Topic
}

func (m *KafkaMessage) Data() (map[string]any, error) {
	var jsonData map[string]any
	err := json.Unmarshal(m.message.Value, &jsonData)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (m *KafkaMessage) Commit() {
	m.kafkaConsumer.CommitMessage(m.message)
}
