package writer

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/osmosis-labs/osmosis/v16/hooks/common"
)

type KafkaWriter struct {
	writer *kafka.Writer // Main Kafka writer instance
}

var _ Writer = &KafkaWriter{}

func NewKafkaWriter(url string) *KafkaWriter {
	paths := strings.SplitN(url, "@", 2)
	return &KafkaWriter{
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:      paths[1:],
			Topic:        paths[0],
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 1 * time.Millisecond,
			BatchBytes:   512000000,
		}),
	}
}

func (k KafkaWriter) WriteMessages(ctx context.Context, msgs []common.Message) {
	kafkaMsgs := make([]kafka.Message, len(msgs))
	for idx, msg := range msgs {
		res, _ := json.Marshal(msg.Value) // Error must always be nil.
		kafkaMsgs[idx] = kafka.Message{Key: []byte(msg.Key), Value: res}
	}
	err := k.writer.WriteMessages(ctx, kafkaMsgs...)
	if err != nil {
		panic(err)
	}
}
