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
	Topics           KafkaTopics `mapstructure:"topics"`
	GroupID          string      `mapstructure:"group_id"`
}

type KafkaTopics struct {
	SessionStarted  string `mapstructure:"session_started"`
	UserRegistered  string `mapstructure:"user_registered"`
	UserLoggedOut   string `mapstructure:"user_logged_out"`
	UserDeleteTopic string `mapstructure:"user_delete_topic"`
}
type SessionServiceConfig struct {
	GrpcAddress string `mapstructure:"grpc_address"`
}
