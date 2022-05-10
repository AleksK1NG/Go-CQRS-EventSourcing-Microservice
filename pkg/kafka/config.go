package kafka

// Config kafka config
type Config struct {
	Brokers    []string `mapstructure:"brokers" validate:"required"`
	GroupID    string   `mapstructure:"groupID" validate:"required,gte=0"`
	InitTopics bool     `mapstructure:"initTopics"`
}

// TopicConfig kafka topic config
type TopicConfig struct {
	TopicName         string `mapstructure:"topicName" validate:"required,gte=0"`
	Partitions        int    `mapstructure:"partitions" validate:"required"`
	ReplicationFactor int    `mapstructure:"replicationFactor" validate:"required"`
}
