package logs

import (
	"log"
	"strings"

	"github.com/niktin06sash/MicroserviceProject/Kafka_service/internal/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	ZapLogger *zap.Logger
}

func (logg *Logger) Sync() {
	logg.ZapLogger.Info("----------CLOSE SERVICE----------")
	logg.ZapLogger.Warn("----------CLOSE SERVICE----------")
	logg.ZapLogger.Error("----------CLOSE SERVICE----------")
	logg.ZapLogger.Sync()
	log.Println("[INFO] [Kafka-Service] Successful sync Logger")
}
func NewLogger(config configs.LoggerConfig) *Logger {
	c := []zapcore.Core{}
	for lvl, fn := range config.Files {
		zapLevel, err := zapcore.ParseLevel(strings.ToUpper(lvl))
		if err != nil {
			log.Fatalf("[ERROR] [Kafka-Service] Error getting the logging level: %v", err)
			return nil
		}
		writer := &lumberjack.Logger{
			Filename:   fn,
			MaxSize:    config.Rotation.MaxSize,
			MaxAge:     config.Rotation.MaxAge,
			MaxBackups: config.Rotation.MaxBackups,
			Compress:   config.Rotation.Compress,
		}
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder := zapcore.NewJSONEncoder(encoderConfig)
		levelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level == zapLevel
		})
		core := zapcore.NewCore(encoder, zapcore.AddSync(writer), levelEnabler)
		c = append(c, core)
	}
	combinedCore := zapcore.NewTee(c...)
	logger := zap.New(combinedCore, zap.AddStacktrace(zapcore.ErrorLevel))
	log.Println("[INFO] [Kafka-Service] Successful connect to Logger")
	logger.Info("----------CLOSE SERVICE----------")
	logger.Warn("----------CLOSE SERVICE----------")
	logger.Error("----------CLOSE SERVICE----------")
	return &Logger{
		ZapLogger: logger,
	}
}
