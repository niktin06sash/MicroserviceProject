package logs

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Logs_service/internal/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	mux     sync.RWMutex
	cores   map[string]zapcore.Core
	encoder zapcore.Encoder
}

func NewLogger(config configs.LoggerConfig, topics configs.KafkaTopics) (*Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	mapa := config.GetTopicFileMap(topics)
	l := &Logger{cores: make(map[string]zapcore.Core), encoder: encoder}
	for topic, filepath := range mapa {
		file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("[DEBUG] [Logs-Service] Error opening log file: %v", err)
			return nil, err
		}
		level := strings.Split(topic, "-")[1]
		zapLevel, err := zapcore.ParseLevel(strings.ToUpper(level))
		if err != nil {
			log.Printf("[DEBUG] [Logs-Service] Error getting the logging level: %v", err)
			return nil, err
		}
		levelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level == zapLevel
		})
		core := zapcore.NewCore(l.encoder, zapcore.AddSync(file), levelEnabler)
		l.cores[topic] = core
	}
	log.Println("[DEBUG] [Logs-Service] Successful connect to Logger")
	return l, nil
}

func (l *Logger) Log(topic string, message string) error {
	l.mux.RLock()
	core, exists := l.cores[topic]
	l.mux.RUnlock()
	if !exists {
		return fmt.Errorf("logger for topic %s not initialized", topic)
	}
	parts := strings.Split(topic, "-")
	if len(parts) < 2 {
		return fmt.Errorf("invalid topic format: %s", topic)
	}
	level := parts[1]
	zapLevel, err := zapcore.ParseLevel(strings.ToUpper(level))
	if err != nil {
		return fmt.Errorf("invalid log level %s: %w", level, err)
	}
	entry := zapcore.Entry{Level: zapLevel, Time: time.Now(), Message: message}
	core.Write(entry, nil)
	return nil
}
func (m *Logger) Sync() {
	m.mux.Lock()
	defer m.mux.Unlock()
	for _, core := range m.cores {
		if closer, ok := core.(io.Closer); ok {
			_ = closer.Close()
		}
	}
	log.Println("[DEBUG] [Logs-Service] Successful close Logger")
}
