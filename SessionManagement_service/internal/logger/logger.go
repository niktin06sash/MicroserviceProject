package logger

import "go.uber.org/zap"

type SessionLoggerInterface interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
}
type SessionLogger struct {
	logger *zap.Logger
}

func NewSessionLogger() *SessionLogger {
	config := zap.NewProductionConfig()
	config.DisableStacktrace = true
	zaplogger, _ := config.Build()
	return &SessionLogger{logger: zaplogger}
}
func (l *SessionLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}
func (l *SessionLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}
func (l *SessionLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}
func (l *SessionLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}
