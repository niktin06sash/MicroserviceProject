package configs

import (
	"log"

	"github.com/spf13/viper"
)

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
}
type KafkaTopics struct {
	InfoLog  string `mapstructure:"info_log"`
	ErrorLog string `mapstructure:"error_log"`
	WarnLog  string `mapstructure:"warn_log"`
}

func LoadConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("internal/configs")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("[ERROR] [Session-Service] Config file not found; using defaults or environment variables")
		} else {
			log.Fatalf("[ERROR] [Session-Service] Error reading config file: %s", err)
		}
	}
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("[ERROR] [Session-Service] Unable to decode into struct, %v", err)
	}
	log.Println("[INFO] [Session-Service] Successful Load Config")
	return config
}
