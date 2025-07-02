package kafka

import (
	"os"
	"strings"

	"github.com/segmentio/kafka-go"
)

func NewProducer() *kafka.Writer {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka:9092"
	}

	return &kafka.Writer{
		Addr:     kafka.TCP(strings.Split(brokers, ",")...),
		Balancer: &kafka.LeastBytes{},
	}
}
