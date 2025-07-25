package configs

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	Database       DatabaseConfig       `mapstructure:"database"`
	Kafka          KafkaConfig          `mapstructure:"kafka"`
	SessionService SessionServiceConfig `mapstructure:"session_service"`
	Redis          RedisConfig          `mapstructure:"redis"`
	RabbitMQ       RabbitMQConfig       `mapstructure:"rabbitmq"`
}

type ServerConfig struct {
	Port             string        `mapstructure:"port"`
	MaxHeaderBytes   int           `mapstructure:"max_header_bytes"`
	ReadTimeout      time.Duration `mapstructure:"read_timeout"`
	WriteTimeout     time.Duration `mapstructure:"write_timeout"`
	IdleTimeout      time.Duration `mapstructure:"idle_timeout"`
	GracefulShutdown time.Duration `mapstructure:"graceful_shutdown"`
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
	BootstrapServers string            `mapstructure:"bootstrap_servers"`
	RetryBackoffMs   int               `mapstructure:"retry_backoff_ms"`
	BatchSize        int               `mapstructure:"batch_size"`
	Acks             string            `mapstructure:"acks"`
	Topics           map[string]string `mapstructure:"topics"`
}
type SessionServiceConfig struct {
	GrpcAddress string `mapstructure:"grpc_address"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}
type RabbitMQConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Name       string `mapstructure:"name"`
	Password   string `mapstructure:"password"`
	Exchange   string `mapstructure:"exchange"`
	RoutingKey string `mapstructure:"routing_key"`
}

func LoadConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("internal/configs")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("[DEBUG] [User-Service] Config file not found; using defaults or environment variables")
		} else {
			log.Fatalf("[DEBUG] [User-Service] Error reading config file: %s", err)
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
		log.Println("[DEBUG] [User-Service] Successful Load Config (docker)")
		return config
	}
	log.Println("[DEBUG] [User-Service] Successful Load Config (localhost)")
	return config
}
func LoadDockerConfig(config *Config) {
	redis := os.Getenv("REDIS_HOST")
	db := os.Getenv("DB_HOST")
	kafka := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	grpcaddr := os.Getenv("SESSION_SERVICE_GRPC_ADDRESS")
	config.Redis.Host = redis
	config.Database.Host = db
	config.Kafka.BootstrapServers = kafka
	config.SessionService.GrpcAddress = grpcaddr
}
