package logs

import (
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap/zapcore"
)

func (l *Logger) NewLog(msg kafka.Message) {
	select {
	case <-l.ctx.Done():
		return
	case l.logchan <- msg:
	default:
		log.Printf("[WARN] [Logs-Service] Log channel is full, dropping log: %+v", msg)
	}
}

func (l *Logger) logWorker() {
	defer l.wg.Done()
	for {
		select {
		case <-l.ctx.Done():
			return
		case msg, ok := <-l.logchan:
			if !ok {
				return
			}
			l.semaphore <- struct{}{}
			defer func() { <-l.semaphore }()
			topic, val := msg.Topic, msg.Value
			l.mux.RLock()
			core, exists := l.cores[topic]
			l.mux.RUnlock()
			if !exists {
				log.Printf("[ERROR] [Logs-Service] Logger for topic %s not initialized", topic)
				continue
			}
			parts := strings.Split(topic, "-")
			if len(parts) < 2 {
				log.Printf("[ERROR] [Logs-Service] Invalid topic format: %s", topic)
				continue
			}
			level := parts[1]
			zapLevel, err := zapcore.ParseLevel(strings.ToUpper(level))
			if err != nil {
				log.Printf("[ERROR] [Logs-Service] Invalid log level: %s: %v", level, err)
				continue
			}
			entry := zapcore.Entry{Level: zapLevel, Time: time.Now(), Message: string(val)}
			core.Write(entry, nil)
			l.store.LogElastic(l.ctx, string(val), level)
		}
	}
}
