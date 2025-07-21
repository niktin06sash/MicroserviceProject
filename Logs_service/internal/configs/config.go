package configs

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Kafka   KafkaConfig   `mapstructure:"kafka"`
	Logger  LoggerConfig  `mapstructure:"logger"`
	Elastic ElasticConfig `mapstruture:"elasticsearch"`
}
type ElasticConfig struct {
	Host  string `mapstructure:"host"`
	Index string `mapstructure:"index"`
}
type KafkaConfig struct {
	BootstrapServers string        `mapstructure:"bootstrap_servers"`
	RetryBackoffMs   int           `mapstructure:"retry_backoff_ms"`
	GroupId          string        `mapstructure:"group_id"`
	Topics           KafkaTopics   `mapstructure:"topics"`
	AutoCommit       bool          `mapstructure:"enable_auto_commit"`
	SessionTimeout   time.Duration `mapstructure:"session_timeout"`
	HearbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
}
type KafkaTopics struct {
	IApi string `mapstructure:"api_info_log"`
	EApi string `mapstructure:"api_error_log"`
	WApi string `mapstructure:"api_warn_log"`

	IUser string `mapstructure:"user_info_log"`
	EUser string `mapstructure:"user_error_log"`
	WUser string `mapstructure:"user_warn_log"`

	ISess string `mapstructure:"session_info_log"`
	ESess string `mapstructure:"session_error_log"`
	WSess string `mapstructure:"session_warn_log"`

	IPhoto string `mapstructure:"photo_info_log"`
	EPhoto string `mapstructure:"photo_error_log"`
	WPhoto string `mapstructure:"photo_warn_log"`
}

func (c LoggerConfig) GetTopicFileMap(Topics KafkaTopics) map[string]string {
	return map[string]string{
		Topics.IApi: c.Files["api_info"],
		Topics.EApi: c.Files["api_error"],
		Topics.WApi: c.Files["api_warn"],

		Topics.IUser: c.Files["user_info"],
		Topics.EUser: c.Files["user_error"],
		Topics.WUser: c.Files["user_warn"],

		Topics.ISess: c.Files["session_info"],
		Topics.ESess: c.Files["session_error"],
		Topics.WSess: c.Files["session_warn"],

		Topics.IPhoto: c.Files["photo_info"],
		Topics.EPhoto: c.Files["photo_error"],
		Topics.WPhoto: c.Files["photo_warn"],
	}
}
func (c KafkaConfig) GetAllTopics() []string {
	return []string{
		c.Topics.EApi,
		c.Topics.WApi,
		c.Topics.IApi,
		c.Topics.WUser,
		c.Topics.IUser,
		c.Topics.EUser,
		c.Topics.WSess,
		c.Topics.ISess,
		c.Topics.ESess,
		c.Topics.EPhoto,
		c.Topics.WPhoto,
		c.Topics.IPhoto,
	}
}

type LoggerConfig struct {
	Files    map[string]string `mapstructure:"files"`
	Rotation struct {
		MaxSize    int  `mapstructure:"max_size"`
		MaxBackups int  `mapstructure:"max_backups"`
		MaxAge     int  `mapstructure:"max_age"`
		Compress   bool `mapstructure:"compress"`
	} `mapstructure:"rotation"`
	Format struct {
		TimeFormat string `mapstructure:"time_format"`
	} `mapstructure:"format"`
}

func LoadConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("internal/configs")

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("[DEBUG] [Logs-Service] Config file not found; using defaults or environment variables")
		} else {
			log.Fatalf("[DEBUG]] [Logs-Service] Error reading config file: %s", err)
		}
	}
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("[DEBUG] [Logs-Service] Unable to decode into struct, %v", err)
	}
	docker_flag := os.Getenv("DOCKER")
	if docker_flag == "TRUE" {
		LoadDockerConfig(&config)
		log.Println("[DEBUG] [Logs-Service] Successful Load Config (docker)")
		return config
	}
	log.Println("[DEBUG] [Logs-Service] Successful Load Config (localhost)")
	return config
}
func LoadDockerConfig(config *Config) {
	kafka := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	el := os.Getenv("ELASTICSEARCH_HOST")
	config.Kafka.BootstrapServers = kafka
	config.Elastic.Host = el
}
