package logs

import (
	"log"

	"github.com/niktin06sash/MicroserviceProject/Kafka_service/internal/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	ZapLogger *zap.Logger
}

func (logg *Logger) Sync() {
	lvl := logg.ZapLogger.Level().String()
	log.Printf("[INFO] [Logger:%s] Successful sync Logger", lvl)
	logg.ZapLogger.Sync()
}
func NewLogger(config configs.LoggerConfig, level string) *Logger {
	filename := config.Files[level]
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		log.Fatalf("[ERROR] [Logger:%s] Error getting the logging level: %v", zapLevel.String(), err)
		return nil
	}
	writer := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    config.Rotation.MaxSize,
		MaxAge:     config.Rotation.MaxAge,
		MaxBackups: config.Rotation.MaxBackups,
		Compress:   config.Rotation.Compress,
	}
	encoderconfig := zap.NewProductionEncoderConfig()
	encoderconfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderconfig)
	core := zapcore.NewCore(encoder, zapcore.AddSync(writer), zapLevel)
	zapLogger := zap.New(core, zap.AddStacktrace(zapcore.ErrorLevel))
	return &Logger{
		ZapLogger: zapLogger,
	}
}
