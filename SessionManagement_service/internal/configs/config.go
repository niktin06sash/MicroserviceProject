package configs

import (
	"log"
	"os"

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
	BootstrapServers string            `mapstructure:"bootstrap_servers"`
	RetryBackoffMs   int               `mapstructure:"retry_backoff_ms"`
	BatchSize        int               `mapstructure:"batch_size"`
	Acks             string            `mapstructure:"acks"`
	Topics           map[string]string `mapstructure:"topics"`
}

func LoadConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("internal/configs")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("[DEBUG] [Session-Service] Config file not found; using defaults or environment variables")
		} else {
			log.Fatalf("[DEBUG] [Session-Service] Error reading config file: %s", err)
		}
	}
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("[DEBUG] [Session-Service] Unable to decode into struct, %v", err)
	}
	docker_flag := os.Getenv("DOCKER")
	if docker_flag == "TRUE" {
		LoadDockerConfig(&config)
		log.Println("[DEBUG] [Session-Service] Successful Load Config (docker)")
		return config
	}
	log.Println("[DEBUG] [Session-Service] Successful Load Config (localhost)")
	return config
}
func LoadDockerConfig(config *Config) {
	redis := os.Getenv("REDIS_HOST")
	kafka := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	config.Redis.Host = redis
	config.Kafka.BootstrapServers = kafka
}
