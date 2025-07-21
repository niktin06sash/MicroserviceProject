package elastic

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

func (el *ElasticClient) NewStorageLog(msg kafka.Message) {
	select {
	case <-el.ctx.Done():
		return
	case el.logchan <- msg:
	default:
		log.Printf("[WARN] [Logs-Service] Log channel is full, dropping log: %+v", msg)
	}
}
func (el *ElasticClient) elasticWorker() {
	defer el.wg.Done()
	for {
		select {
		case <-el.ctx.Done():
			return
		case msg, ok := <-el.logchan:
			if !ok {
				return
			}
			el.semaphor <- struct{}{}
			logEntry := map[string]interface{}{
				"@timestamp": time.Now().UTC().Format(time.RFC3339),
				"level":      strings.Split(msg.Topic, "-")[1],
				"message":    string(msg.Value),
			}
			data, err := json.Marshal(logEntry)
			if err != nil {
				log.Printf("[WARN] [Logs-Service] Log marshal Error: %+v", msg)
			}
			_, err = el.client.Index(el.index, bytes.NewReader(data))
			if err != nil {
				log.Printf("[WARN] [Logs-Service] Error while save log in Elasticsearch: %+v", msg)
			}
			<-el.semaphor
		}
	}
}
