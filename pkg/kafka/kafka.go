package kafka

import (
	"strings"

	kafkago "github.com/segmentio/kafka-go"
)

type WriterConfig struct {
	Brokers              []string
	Topic                string
	AllowAutoTopicCreate bool
}

func ParseBrokers(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func NewWriter(cfg WriterConfig) *kafkago.Writer {
	return &kafkago.Writer{
		Addr:                   kafkago.TCP(cfg.Brokers...),
		Topic:                  cfg.Topic,
		Balancer:               &kafkago.LeastBytes{},
		AllowAutoTopicCreation: cfg.AllowAutoTopicCreate,
	}
}
