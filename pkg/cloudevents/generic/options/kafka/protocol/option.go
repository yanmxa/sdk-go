package kafka_confluent

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

//TODO: the implementation will be removed once the this pr is merged: https://github.com/cloudevents/sdk-go/pull/988

const (
	OffsetEventSource = "io.cloudevents.kafka.confluent.consumer"
	OffsetEventType   = "io.cloudevents.kafka.confluent.consumer.offsets"
)

// Option is the function signature required to be considered an kafka_confluent.Option.
type Option func(*Protocol) error

// WithConfigMap sets the configMap to init the kafka client. This option is not required.
func WithConfigMap(config *kafka.ConfigMap) Option {
	return func(p *Protocol) error {
		if config == nil {
			return fmt.Errorf("the kafka.ConfigMap option must not be nil")
		}
		p.kafkaConfigMap = config
		return nil
	}
}

// WithSenderTopic sets the defaultTopic for the kafka.Producer. This option is not required.
func WithSenderTopic(defaultTopic string) Option {
	return func(p *Protocol) error {
		if defaultTopic == "" {
			return fmt.Errorf("the producer topic option must not be nil")
		}
		p.producerDefaultTopic = defaultTopic
		return nil
	}
}

// WithDeliveryChan sets the deliveryChan for the kafka.Producer. This option is not required.
func WithDeliveryChan(deliveryChan chan kafka.Event) Option {
	return func(p *Protocol) error {
		if deliveryChan == nil {
			return fmt.Errorf("the producer deliveryChan option must not be nil")
		}
		p.producerDeliveryChan = deliveryChan
		return nil
	}
}

// WithReceiverTopics sets the topics for the kafka.Consumer. This option is not required.
func WithReceiverTopics(topics []string) Option {
	return func(p *Protocol) error {
		if topics == nil {
			return fmt.Errorf("the consumer topics option must not be nil")
		}
		p.consumerTopics = topics
		return nil
	}
}

// WithRebalanceCallBack sets the callback for rebalancing of the consumer group. This option is not required.
func WithRebalanceCallBack(rebalanceCb kafka.RebalanceCb) Option {
	return func(p *Protocol) error {
		if rebalanceCb == nil {
			return fmt.Errorf("the consumer group rebalance callback must not be nil")
		}
		p.consumerRebalanceCb = rebalanceCb
		return nil
	}
}

// WithPollTimeout sets timeout of the consumer polling for message or events, return nil on timeout. This option is not required.
func WithPollTimeout(timeoutMs int) Option {
	return func(p *Protocol) error {
		p.consumerPollTimeout = timeoutMs
		return nil
	}
}

// WithSender set a kafka.Producer instance to init the client directly. This option is not required.
func WithSender(producer *kafka.Producer) Option {
	return func(p *Protocol) error {
		if producer == nil {
			return fmt.Errorf("the producer option must not be nil")
		}
		p.producer = producer
		return nil
	}
}

// WithErrorHandler provide a func on how to handle the kafka.Error which the kafka.Consumer has polled. This option is not required.
func WithErrorHandler(handler func(err kafka.Error)) Option {
	return func(p *Protocol) error {
		p.consumerErrorHandler = handler
		return nil
	}
}

// WithSender set a kafka.Consumer instance to init the client directly. This option is not required.
func WithReceiver(consumer *kafka.Consumer) Option {
	return func(p *Protocol) error {
		if consumer == nil {
			return fmt.Errorf("the consumer option must not be nil")
		}
		p.consumer = consumer
		return nil
	}
}

// Opaque key type used to store topicPartitionOffsets: assign them from ctx. This option is not required.
type topicPartitionOffsetsType struct{}

var offsetKey = topicPartitionOffsetsType{}

// WithTopicPartitionOffsets will set the positions where the consumer starts consuming from. This option is not required.
func WithTopicPartitionOffsets(ctx context.Context, topicPartitionOffsets []kafka.TopicPartition) context.Context {
	if len(topicPartitionOffsets) == 0 {
		panic("the topicPartitionOffsets cannot be empty")
	}
	for _, offset := range topicPartitionOffsets {
		if offset.Topic == nil || *(offset.Topic) == "" {
			panic("the kafka topic cannot be nil or empty")
		}
		if offset.Partition < 0 || offset.Offset < 0 {
			panic("the kafka partition/offset must be non-negative")
		}
	}
	return context.WithValue(ctx, offsetKey, topicPartitionOffsets)
}

// TopicPartitionOffsetsFrom looks in the given context and returns []kafka.TopicPartition or nil if not set
func TopicPartitionOffsetsFrom(ctx context.Context) []kafka.TopicPartition {
	c := ctx.Value(offsetKey)
	if c != nil {
		if s, ok := c.([]kafka.TopicPartition); ok {
			return s
		}
	}
	return nil
}

// Opaque key type used to store message key
type messageKeyType struct{}

var keyForMessageKey = messageKeyType{}

// WithMessageKey returns back a new context with the given messageKey. This option is not required.
func WithMessageKey(ctx context.Context, messageKey string) context.Context {
	return context.WithValue(ctx, keyForMessageKey, messageKey)
}

// MessageKeyFrom looks in the given context and returns `messageKey` as a string if found and valid, otherwise "".
func MessageKeyFrom(ctx context.Context) string {
	c := ctx.Value(keyForMessageKey)
	if c != nil {
		if s, ok := c.(string); ok {
			return s
		}
	}
	return ""
}
