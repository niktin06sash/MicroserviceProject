package configs

type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	SessionService SessionServiceConfig `mapstructure:"session_service"`
	Kafka          KafkaConfig          `mapstructure:"kafka"`
	Routes         map[string]string    `mapstructure:"routes"`
	SSL            SSLConfig            `mapstructure:"ssl"`
}
type ServerConfig struct {
	Port string `mapstructure:"port"`
}
type SessionServiceConfig struct {
	GrpcAddress string `mapstructure:"grpc_address"`
}
type SSLConfig struct {
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}
type KafkaConfig struct {
	BootstrapServers string      `mapstructure:"bootstrap_servers"`
	RetryBackoffMs   int         `mapstructure:"retry_backoff_ms"`
	BatchSize        int         `mapstructure:"batch_size"`
	Acks             string      `mapstructure:"acks"`
	Topics           KafkaTopics `mapstructure:"topics"`
}
type KafkaTopics struct {
	InfoLog  string `mapstructure:"api-info-log-topic"`
	ErrorLog string `mapstructure:"api-error-log-topic"`
	WarnLog  string `mapstructure:"api-warn-log-topic"`
}
