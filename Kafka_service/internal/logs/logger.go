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
	Writers   []*lumberjack.Logger
}

func (logg *Logger) Sync() {
	logg.ZapLogger.Info("----------CLOSE SERVICE----------")
	logg.ZapLogger.Warn("----------CLOSE SERVICE----------")
	logg.ZapLogger.Error("----------CLOSE SERVICE----------")
	logg.ZapLogger.Sync()
	for _, writer := range logg.Writers {
		writer.Close()
	}
	log.Println("[INFO] [Kafka-Service] Successful sync Logger")
}
func NewLogger(config configs.LoggerConfig) *Logger {
	c := []zapcore.Core{}
	writers := []*lumberjack.Logger{}
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
		writers = append(writers, writer)
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
	logger.Info("----------START SERVICE----------")
	logger.Warn("----------START SERVICE----------")
	logger.Error("----------START SERVICE----------")
	return &Logger{
		ZapLogger: logger,
		Writers:   writers,
	}
}
