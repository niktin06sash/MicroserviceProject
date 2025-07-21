package logs

import (
	"context"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/niktin06sash/MicroserviceProject/Logs_service/internal/configs"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	mux       sync.RWMutex
	cores     map[string]zapcore.Core
	encoder   zapcore.Encoder
	store     StorageLog
	ctx       context.Context
	cancel    context.CancelFunc
	logchan   chan kafka.Message
	semaphore chan struct{}
	wg        *sync.WaitGroup
}
type StorageLog interface {
	LogElastic(ctx context.Context, val string, level string)
}

func NewLogger(config configs.LoggerConfig, topics configs.KafkaTopics, store StorageLog) (*Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	mapa := config.GetTopicFileMap(topics)
	ctx, cancel := context.WithCancel(context.Background())
	l := &Logger{cores: make(map[string]zapcore.Core),
		encoder: encoder, store: store, ctx: ctx, cancel: cancel, logchan: make(chan kafka.Message, 10000), semaphore: make(chan struct{}, 10), wg: &sync.WaitGroup{}}
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
	for i := 0; i < 100; i++ {
		l.wg.Add(1)
		go l.logWorker()
	}
	log.Println("[DEBUG] [Logs-Service] Successful connect to Logger")
	return l, nil
}

func (l *Logger) Sync() {
	l.cancel()
	close(l.logchan)
	l.wg.Wait()
	l.mux.Lock()
	defer l.mux.Unlock()
	for _, core := range l.cores {
		if closer, ok := core.(io.Closer); ok {
			_ = closer.Close()
		}
	}
	log.Println("[DEBUG] [Logs-Service] Successful close Logger")
}
