package kafka

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"time"
)

func (kf *KafkaConsumerGroup) readLogs() {
	defer kf.wg.Done()
	for {
		select {
		case <-kf.ctx.Done():
			log.Println("[DEBUG] [Logs-Service] Context canceled, stopping working KafkaConsumer-Group...")
			return
		default:
			ctx, cancel := context.WithTimeout(kf.ctx, 5*time.Second)
			defer cancel()
			msg, err := kf.reader.FetchMessage(ctx)
			if err != nil {
				if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
					log.Printf("[ERROR] [Logs-Service] Failed to read log: %v", err)
				}
				continue
			}
			atomic.AddInt64(&kf.counter, 1)
			if err := kf.logger.Log(msg.Topic, string(msg.Value)); err != nil {
				log.Printf("[ERROR] [Logs-Service] Failed to send log from topic %s: %v", msg.Topic, err)
			}
			if err := kf.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("[ERROR] [Logs-Service] Failed to commit offset: %v", err)
			}
			cancel()
		}
	}
}
