package searcher

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"time"
)

func (el *ElasticClient) LogElastic(ctx context.Context, val string, level string) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	logEntry := map[string]interface{}{
		"@timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":      level,
		"message":    val,
	}
	data, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("[WARN] [Logs-Service] Log marshal Error: %+v", val)
		return
	}
	_, err = el.client.Index(el.index, bytes.NewReader(data))
	if err != nil {
		log.Printf("[WARN] [Logs-Service] Error while save log in Elasticsearch: %+v", val)
		return
	}
}
