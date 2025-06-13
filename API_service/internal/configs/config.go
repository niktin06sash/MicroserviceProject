package configs

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	SessionService SessionServiceConfig `mapstructure:"session_service"`
	PhotoService   PhotoServiceConfig   `mapstructure:"photo_service"`
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
type PhotoServiceConfig struct {
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

func LoadConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("internal/configs")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("[DEBUG] [API-Service] Config file not found; using defaults or environment variables")
		} else {
			log.Fatalf("[DEBUG] [API-Service] Error reading config file: %s", err)
		}
	}
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("[DEBUG] [API-Service] Unable to decode into struct, %v", err)
	}
	docker_flag := os.Getenv("DOCKER")
	if docker_flag == "TRUE" {
		LoadDockerConfig(&config)
		log.Println("[DEBUG] [API-Service] Successful Load Config (docker)")
		return config
	}
	log.Println("[DEBUG] [API-Service] Successful Load Config (localhost)")
	return config
}
func LoadDockerConfig(config *Config) {
	sessionaddr := os.Getenv("SESSION_SERVICE_GRPC_ADDRESS")
	photoaddr := os.Getenv("PHOTO_SERVICE_GRPC_ADDRESS")
	userservice := os.Getenv("USER_SERVICE_HOST")
	kafka := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	config.SessionService.GrpcAddress = sessionaddr
	config.PhotoService.GrpcAddress = photoaddr
	config.Kafka.BootstrapServers = kafka
	config.Routes = UpdateRoutes(config.Routes, userservice)
}
func UpdateRoutes(routes map[string]string, host string) map[string]string {
	updatedRoutes := make(map[string]string)
	for path, url := range routes {
		updatedRoutes[path] = replaceHost(url, host)
	}
	return updatedRoutes
}

func replaceHost(url, newHost string) string {
	parts := strings.SplitN(url, "://", 2)
	if len(parts) != 2 {
		return url
	}
	protocol := parts[0]
	address := parts[1]
	hostAndPath := strings.SplitN(address, "/", 2)
	if len(hostAndPath) < 1 {
		return url
	}
	newAddress := fmt.Sprintf("%s/%s", newHost, hostAndPath[1])
	return fmt.Sprintf("%s://%s", protocol, newAddress)
}
