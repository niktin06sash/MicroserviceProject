package configs

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Kafka  KafkaConfig  `mapstructure:"kafka"`
	Redis  RedisConfig  `mapstructure:"redis"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}
type KafkaConfig struct {
	BootstrapServers string      `mapstructure:"bootstrap_servers"`
	RetryBackoffMs   int         `mapstructure:"retry_backoff_ms"`
	BatchSize        int         `mapstructure:"batch_size"`
	Acks             string      `mapstructure:"acks"`
	Topics           KafkaTopics `mapstructure:"topics"`
	GroupID          string      `mapstructure:"group_id"`
}
type KafkaTopics struct {
	InfoLog  string `mapstructure:"session-info-log-topic"`
	ErrorLog string `mapstructure:"session-error-log-topic"`
	WarnLog  string `mapstructure:"session-warn-log-topic"`
}
