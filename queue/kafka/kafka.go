package kafka

import (
	kafkago "github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewWriter(c *Config) *kafkago.Writer {
	return kafkago.NewWriter(kafkago.WriterConfig{
		Brokers: c.Brokers,
		Topic:   c.Topic,
	})
}

func NewReader(c *Config) *kafkago.Reader {
	return kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: c.Brokers,
		Topic:   c.Topic,
		GroupID: c.GroupID,
	})
}
