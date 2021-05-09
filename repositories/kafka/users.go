package kafka

import (
	"code/tech-test/domain/users/models"
	jsonSerializer "code/tech-test/repositories/json"
	"encoding/json"
	"fmt"

	kafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

type UserSerializer interface {
	SerializeUser(user models.User) jsonSerializer.UserMessage
}

type UserProducer struct {
	serializer UserSerializer
	producer   *kafka.Producer
	topic      string
}

func NewUserProducer(producer *kafka.Producer, topic string, userSerializer UserSerializer) UserProducer {
	return UserProducer{
		producer:   producer,
		topic:      topic,
		serializer: userSerializer,
	}
}

func (p UserProducer) Publish(user models.User) error {

	message, err := json.Marshal(p.serializer.SerializeUser(user))
	if err != nil {
		return fmt.Errorf("%w failed to marshal message", err)
	}

	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, nil)
	if err != nil {
		return fmt.Errorf("%w failed to publish message", err)
	}

	return nil
}
