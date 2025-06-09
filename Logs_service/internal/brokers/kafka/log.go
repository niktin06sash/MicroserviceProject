package kafka

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

func (kf *KafkaConsumer) readLogs() {
	defer kf.wg.Done()
	for {
		select {
		case <-kf.ctx.Done():
			return
		default:
			ctx, cancel := context.WithTimeout(kf.ctx, 2*time.Second)
			defer cancel()
			msg, err := kf.reader.FetchMessage(ctx)
			if err != nil {
				if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
					log.Printf("[ERROR] [Logs-Service] [KafkaConsumer:%s] Failed to read log: %v", kf.reader.Config().Topic, err)
				}
				continue
			}
			atomic.AddInt64(&kf.counter, 1)
			parts := strings.Split(kf.reader.Config().Topic, "-")
			level := parts[1]
			switch level {
			case "info":
				kf.logger.ZapLogger.Info(string(msg.Value), zap.Int64("number", kf.counter))
			case "error":
				kf.logger.ZapLogger.Error(string(msg.Value), zap.Int64("number", kf.counter))
			case "warn":
				kf.logger.ZapLogger.Warn(string(msg.Value), zap.Int64("number", kf.counter))
			}
			if err := kf.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("[ERROR] [Logs-Service] [KafkaConsumer:%s] Failed to commit offset: %v", kf.reader.Config().Topic, err)
			}
		}
	}
}
