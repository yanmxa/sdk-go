package kafka

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	kafka_confluent "open-cluster-management.io/sdk-go/pkg/cloudevents/generic/options/kafka/protocol"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/types"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	defaultSpecTopic   = "spec"
	defaultStatusTopic = "status"
)

type KafkaOptions struct {
	// the configMap: https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
	ConfigMap *kafka.ConfigMap `json:"configs,omitempty" yaml:"configs,omitempty"`
	Topics    *types.Topics    `json:"topics,omitempty" yaml:"topics,omitempty"`
}

func NewKafkaOptions() *KafkaOptions {
	return &KafkaOptions{
		Topics: &types.Topics{
			SourceEvents: defaultSpecTopic,
			AgentEvents:  defaultStatusTopic,
		},
	}
}

func (o *KafkaOptions) GetCloudEventsClient(clientOpts ...kafka_confluent.Option) (cloudevents.Client, error) {
	protocol, err := kafka_confluent.New(clientOpts...)
	if err != nil {
		return nil, err
	}
	return cloudevents.NewClient(protocol)
}

// BuildKafkaOptionsFromFlags builds configs from a config filepath.
func BuildKafkaOptionsFromFlags(configPath string) (*KafkaOptions, error) {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var opts KafkaOptions
	if err := yaml.Unmarshal(configData, &opts); err != nil {
		return nil, err
	}

	if opts.ConfigMap == nil {
		return nil, fmt.Errorf("the configs should be set")
	}

	val, err := opts.ConfigMap.Get("bootstrap.servers", "")
	if err != nil {
		return nil, err
	}
	if val == "" {
		return nil, fmt.Errorf("bootstrap.servers is required")
	}

	options := &KafkaOptions{
		ConfigMap: opts.ConfigMap,
		Topics: &types.Topics{
			SourceEvents: defaultSpecTopic,
			AgentEvents:  defaultStatusTopic,
		},
	}
	if opts.Topics != nil {
		options.Topics = opts.Topics
	}

	if options.Topics.SourceEvents == "" || options.Topics.AgentEvents == "" {
		return nil, fmt.Errorf("the topic value should be set")
	}
	return options, nil
}
