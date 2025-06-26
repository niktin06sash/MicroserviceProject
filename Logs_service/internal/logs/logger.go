package logs

import (
	"log"
	"os"
	"strings"

	"github.com/niktin06sash/MicroserviceProject/Logs_service/internal/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	ZapLogger *zap.Logger
	File      *os.File
	Level     string
	Topic     string
}

func (logg *Logger) Sync() {
	logg.ZapLogger.Sync()
	logg.File.Close()
	log.Printf("[DEBUG] [Logs-Service] [Logger: %s] Successful sync and close Logger", logg.Topic)
}
func NewLogger(config configs.LoggerConfig, topic string) (*Logger, error) {
	parts := strings.Split(topic, "-")
	service := parts[0]
	level := parts[1]
	fn := config.Files[service+"_"+level]
	zapLevel, err := zapcore.ParseLevel(strings.ToUpper(level))
	if err != nil {
		log.Printf("[DEBUG] [Logs-Service] Error getting the logging level: %v", err)
		return nil, err
	}
	file, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[DEBUG] [Logs-Service] Error opening log file: %v", err)
		return nil, err
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	levelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapLevel
	})
	core := zapcore.NewCore(encoder, zapcore.AddSync(file), levelEnabler)
	logger := zap.New(core)
	log.Printf("[DEBUG] [Logs-Service] [Logger: %s] Successful connect to Logger", topic)
	return &Logger{
		ZapLogger: logger,
		File:      file,
		Level:     level,
		Topic:     topic,
	}, nil
}
