package configs

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Kafka  KafkaConfig  `mapstructure:"kafka"`
	Logger LoggerConfig `mapstructure:"logger"`
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
}

func (c Config) GetAllTopics() []string {
	return []string{
		c.Kafka.Topics.EApi,
		c.Kafka.Topics.WApi,
		c.Kafka.Topics.IApi,
		c.Kafka.Topics.WUser,
		c.Kafka.Topics.IUser,
		c.Kafka.Topics.EUser,
		c.Kafka.Topics.WSess,
		c.Kafka.Topics.ISess,
		c.Kafka.Topics.ESess,
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
			log.Printf("[ERROR] [Kafka-Service] Config file not found; using defaults or environment variables")
		} else {
			log.Fatalf("[ERROR] [Kafka-Service] Error reading config file: %s", err)
		}
	}
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("[ERROR] [Kafka-Service] Unable to decode into struct, %v", err)
	}
	log.Println("[INFO] [Kafka-Service] Successful Load Config")
	return config
}
