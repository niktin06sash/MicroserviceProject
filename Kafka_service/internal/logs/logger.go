package logs

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Kafka_service/internal/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	ZapLogger *zap.Logger
}

func (logg *Logger) Sync() {
	closemsg := fmt.Sprintf("----------CLOSE SERVICE IN %v ----------", time.Now())
	logg.ZapLogger.Info(closemsg)
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
		core := zapcore.NewCore(encoder, zapcore.AddSync(writer), zapLevel)
		c = append(c, core)
	}
	combinedCore := zapcore.NewTee(c...)
	logger := zap.New(combinedCore, zap.AddStacktrace(zapcore.ErrorLevel))
	log.Println("[INFO] [Kafka-Service] Successful connect to Logger")
	startmsg := fmt.Sprintf("----------START SERVICE IN %v ----------", time.Now())
	logger.Info(startmsg)
	return &Logger{
		ZapLogger: logger,
	}
}
