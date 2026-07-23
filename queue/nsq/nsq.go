package nsq

import (
	nsqio "github.com/nsqio/go-nsq"

	"github.com/ao9911/go-matrix/log"
)

type Config struct {
	Address string
	Topic   string
	Channel string
}

func NewProducer(c *Config) *nsqio.Producer {
	producer, err := nsqio.NewProducer(c.Address, nsqio.NewConfig())
	if err != nil {
		log.Fatalf("nsq.NewProducer error: %v", err)
	}
	return producer
}

func NewConsumer(c *Config) *nsqio.Consumer {
	consumer, err := nsqio.NewConsumer(c.Topic, c.Channel, nsqio.NewConfig())
	if err != nil {
		log.Fatalf("nsq.NewConsumer error: %v", err)
	}
	return consumer
}
