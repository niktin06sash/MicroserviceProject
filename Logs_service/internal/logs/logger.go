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
	switch logg.Level {
	case "info":
		logg.ZapLogger.Info("----------CLOSE SERVICE----------")
	case "error":
		logg.ZapLogger.Error("----------CLOSE SERVICE----------")
	case "warn":
		logg.ZapLogger.Warn("----------CLOSE SERVICE----------")
	}
	logg.ZapLogger.Sync()
	logg.File.Close()
	log.Printf("[INFO] [Logs-Service] [Logger: %s] Successful sync and close Logger", logg.Topic)
}
func NewLogger(config configs.LoggerConfig, topic string) *Logger {
	parts := strings.Split(topic, "-")
	service := parts[0]
	level := parts[1]
	fn := config.Files[service+"_"+level]
	zapLevel, err := zapcore.ParseLevel(strings.ToUpper(level))
	if err != nil {
		log.Fatalf("[ERROR] [Logs-Service] Error getting the logging level: %v", err)
		return nil
	}
	file, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		println("[ERROR] [Logs-Service] Error opening log file:", err)
		return nil
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	levelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapLevel
	})
	core := zapcore.NewCore(encoder, zapcore.AddSync(file), levelEnabler)
	logger := zap.New(core)
	log.Printf("[INFO] [Logs-Service] [Logger: %s] Successful connect to Logger", topic)
	switch level {
	case "info":
		logger.Info("----------START SERVICE----------")
	case "error":
		logger.Error("----------START SERVICE----------")
	case "warn":
		logger.Warn("----------START SERVICE----------")
	}
	return &Logger{
		ZapLogger: logger,
		File:      file,
		Level:     level,
		Topic:     topic,
	}
}
