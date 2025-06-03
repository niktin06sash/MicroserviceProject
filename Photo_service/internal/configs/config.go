package configs

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"mongodb"`
	Redis    RedisConfig    `mapstructure:"redis"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Mega     MegaConfig     `mapstructure:"mega"`
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
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type RabbitMQConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Name        string `mapstructure:"name"`
	Password    string `mapstructure:"password"`
	Queue       string `mapstructure:"queue"`
	Exchange    string `mapstructure:"exchange"`
	ConsumerTag string `mapstructure:"consumer_tag"`
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
type MegaConfig struct {
	Email         string `mapstructure:"email"`
	Password      string `mapstructure:"password"`
	MainDirectory string `mapstructure:"main_directory"`
}

func LoadConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("internal/configs")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("[DEBUG] [Photo-Service] Config file not found; using defaults or environment variables")
		} else {
			log.Fatalf("[DEBUG] [Photo-Service] Error reading config file: %s", err)
		}
	}
	viper.SetConfigName("mega")
	viper.SetConfigType("yml")
	viper.AddConfigPath("internal/configs")
	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("[DEBUG] [Photo-Service] Mega file not found; using defaults or environment variables")
		} else {
			log.Fatalf("[DEBUG] [Photo-Service] Error reading mega file: %s", err)
		}
	}
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("[DEBUG] [Photo-Service] Unable to decode into struct, %v", err)
	}
	docker_flag := os.Getenv("DOCKER")
	if docker_flag == "TRUE" {
		LoadDockerConfig(&config)
		log.Println("[DEBUG] [Photo-Service] Successful Load Config (docker)")
		return config
	}
	log.Println("[DEBUG] [Photo-Service] Successful Load Config (localhost)")
	return config
}
func LoadDockerConfig(config *Config) {
	redis := os.Getenv("REDIS_HOST")
	kafka := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	rabbit := os.Getenv("RABBITMQ_HOST")
	db := os.Getenv("DB_HOST")
	config.Redis.Host = redis
	config.Kafka.BootstrapServers = kafka
	config.RabbitMQ.Host = rabbit
	config.Database.Host = db
}
