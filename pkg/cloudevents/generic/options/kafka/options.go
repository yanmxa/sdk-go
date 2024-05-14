//go:build kafka

package kafka

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"

	kafkav2 "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	// sourceEventsTopic is a topic for sources to publish their resource create/update/delete events, the first
	// asterisk is a wildcard for source, the second asterisk is a wildcard for cluster.
	sourceEventsTopic = "sourceevents.*.*"
	// agentEventsTopic is a topic for agents to publish their resource status update events, the first
	// asterisk is a wildcard for source, the second asterisk is a wildcard for cluster.
	agentEventsTopic = "agentevents.*.*"
	// sourceBroadcastTopic is for a source to publish its events to all agents, the asterisk is a wildcard for source.
	sourceBroadcastTopic = "sourcebroadcast.*"
	// agentBroadcastTopic is for a agent to publish its events to all sources, the asterisk is a wildcard for cluster.
	agentBroadcastTopic = "agentbroadcast.*"
)

type KafkaOptions map[string]interface{}

func (opts *KafkaOptions) ConfigMap() kafkav2.ConfigMap {
	kafkaConfigMap := kafkav2.ConfigMap{}
	for k, v := range *opts {
		_ = kafkaConfigMap.SetKey(k, v)
	}
	return kafkaConfigMap
}

type KafkaConfig struct {
	// BootstrapServer is the host of the Kafka broker (hostname:port).
	BootstrapServer string `json:"bootstrapServer" yaml:"bootstrapServer"`

	// CAFile is the file path to a cert file for the MQTT broker certificate authority.
	CAFile string `json:"caFile,omitempty" yaml:"caFile,omitempty"`
	// ClientCertFile is the file path to a client cert file for TLS.
	ClientCertFile string `json:"clientCertFile,omitempty" yaml:"clientCertFile,omitempty"`
	// ClientKeyFile is the file path to a client key file for TLS.
	ClientKeyFile string `json:"clientKeyFile,omitempty" yaml:"clientKeyFile,omitempty"`

	// GroupID is a string that uniquely identifies the group of consumer processes to which this consumer belongs.
	// Each different application will have a unique consumer GroupID. The default value is agentID for agent, sourceID for source
	GroupID string `json:"groupID,omitempty" yaml:"groupID,omitempty"`
}

// Listen to all the events on the default events channel
// It's important to read these events otherwise the events channel will eventually fill up
// Detail: https://github.com/cloudevents/sdk-go/blob/main/protocol/kafka_confluent/v2/protocol.go#L90
func handleProduceEvents(producerEvents chan kafkav2.Event, errChan chan error) {
	if producerEvents == nil {
		return
	}
	go func() {
		for e := range producerEvents {
			switch ev := e.(type) {
			case *kafkav2.Message:
				// The message delivery report, indicating success or failure when sending message
				if ev.TopicPartition.Error != nil {
					klog.Errorf("Delivery failed: %v", ev.TopicPartition.Error)
				}
			case kafkav2.Error:
				// Generic client instance-level errors, such as
				// broker connection failures, authentication issues, etc.
				errChan <- fmt.Errorf("client error %w", ev)
			}
		}
	}()
}

// BuildKafkaOptionsFromFlags builds configs from a config filepath.
func BuildKafkaOptionsFromFlags(configPath string) (*KafkaOptions, error) {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// TODO: failed to unmarshal the data to kafka.ConfigMap directly.
	// Further investigation is required to understand the reasons behind it.
	config := &KafkaConfig{}
	if err := yaml.Unmarshal(configData, config); err != nil {
		return nil, err
	}

	if config.BootstrapServer == "" {
		return nil, fmt.Errorf("bootstrapServer is required")
	}

	if (config.ClientCertFile == "" && config.ClientKeyFile != "") ||
		(config.ClientCertFile != "" && config.ClientKeyFile == "") {
		return nil, fmt.Errorf("either both or none of clientCertFile and clientKeyFile must be set")
	}
	if config.ClientCertFile != "" && config.ClientKeyFile != "" && config.CAFile == "" {
		return nil, fmt.Errorf("setting clientCertFile and clientKeyFile requires caFile")
	}

	kafkaOptions := KafkaOptions{
		"bootstrap.servers":       config.BootstrapServer,
		"socket.keepalive.enable": true,
		// silence spontaneous disconnection logs, kafka recovers by itself.
		"log.connection.close": false,
		// https://github.com/confluentinc/librdkafka/issues/4349
		"ssl.endpoint.identification.algorithm": "none",
		// the events channel server for both producer and consumer
		"go.events.channel.size": 1000,

		// producer
		"acks":    "1",
		"retries": "0",

		// consumer
		"group.id": config.GroupID,

		// If true the consumer’s offset will be periodically committed in the background.
		"enable.auto.commit": true,

		// If true (default) the client will automatically store the offset+1 of the message just prior to passing the message to the application.
		// The offset is stored in memory and will be used by the next call to commit() (without explicit offsets specified) or the next auto commit.
		// If false and enable.auto.commit=true, the application will manually have to call rd_kafka_offset_store() to store the offset to auto commit. (optional)
		"enable.auto.offset.store":   true,
		"queued.max.messages.kbytes": 32768, // 32 MB

		// earliest: automatically reset the offset to the earliest offset
		// latest: automatically reset the offset to the latest offset
		"auto.offset.reset": "latest",

		// The frequency in milliseconds that the consumer offsets are commited (written) to offset storage
		"auto.commit.interval.ms": 5000,
	}

	if config.ClientCertFile != "" {
		kafkaOptions["security.protocol"] = "ssl"
		kafkaOptions["ssl.ca.location"] = config.CAFile
		kafkaOptions["ssl.certificate.location"] = config.ClientCertFile
		kafkaOptions["ssl.key.location"] = config.ClientKeyFile

		// _ = kafkaConfigMap.SetKey("security.protocol", "ssl")
		// _ = kafkaConfigMap.SetKey("ssl.ca.location", config.CAFile)
		// _ = kafkaConfigMap.SetKey("ssl.certificate.location", config.ClientCertFile)
		// _ = kafkaConfigMap.SetKey("ssl.key.location", config.ClientKeyFile)
	}
	return &kafkaOptions, nil
}

// func convertToKafkaConfigMap(configMap map[string]interface{}) kafkav2.ConfigMap {
// 	kafkaConfigMap := kafkav2.ConfigMap{}
// 	for k, v := range configMap {
// 		_ = kafkaConfigMap.SetKey(k, v)
// 	}
// 	return kafkaConfigMap
// }
