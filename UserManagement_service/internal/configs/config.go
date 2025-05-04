package configs

type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	Database       DatabaseConfig       `mapstructure:"database"`
	Kafka          KafkaConfig          `mapstructure:"kafka"`
	SessionService SessionServiceConfig `mapstructure:"session_service"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}
type KafkaConfig struct {
	BootstrapServers string      `mapstructure:"bootstrap_servers"`
	RetryBackoffMs   int         `mapstructure:"retry_backoff_ms"`
	BatchSize        int         `mapstructure:"batch_size"`
	Acks             string      `mapstructure:"acks"`
	Topics           KafkaTopics `mapstructure:"topics"`
}

type KafkaTopics struct {
	InfoLog  string `mapstructure:"info_log"`
	ErrorLog string `mapstructure:"error_log"`
	WarnLog  string `mapstructure:"warn_log"`
}
type SessionServiceConfig struct {
	GrpcAddress string `mapstructure:"grpc_address"`
}
