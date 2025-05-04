package configs

import (
	"log"

	"github.com/spf13/viper"
)

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
	InfoLog  string `mapstructure:"info_log"`
	ErrorLog string `mapstructure:"error_log"`
	WarnLog  string `mapstructure:"warn_log"`
}

func LoadConfig(path string) Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(path)

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("[ERROR] [API-Service] Config file not found; using defaults or environment variables")
		} else {
			log.Fatalf("[ERROR] [API-Service] Error reading config file: %s", err)
		}
	}
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("[ERROR] [API-Service] Unable to decode into struct, %v", err)
	}
	return config
}
