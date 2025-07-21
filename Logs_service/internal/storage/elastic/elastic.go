package elastic

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/niktin06sash/MicroserviceProject/Logs_service/internal/configs"
	"github.com/segmentio/kafka-go"
)

type ElasticClient struct {
	client   *elasticsearch.Client
	logchan  chan kafka.Message
	semaphor chan struct{}
	wg       *sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	index    string
}

func NewElasticClient(c configs.ElasticConfig) (*ElasticClient, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			c.Host,
		},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Printf("[DEBUG] [Logs-Service] Error while connecting to Elasticsearch: %v", err)
		return nil, err
	}
	if err := createIndexIfNotExists(client, c.Index); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	el := &ElasticClient{
		client:   client,
		logchan:  make(chan kafka.Message, 1000),
		ctx:      ctx,
		cancel:   cancel,
		wg:       &sync.WaitGroup{},
		semaphor: make(chan struct{}, 10),
		index:    c.Index,
	}
	for i := 0; i < 100; i++ {
		el.wg.Add(1)
		go el.elasticWorker()
	}
	log.Println("[DEBUG] [Logs-Service] Successful connect to Elasticsearch")
	return el, nil
}

func (ec *ElasticClient) Close() {
	ec.cancel()
	close(ec.logchan)
	ec.wg.Wait()
	log.Println("[DEBUG] [Logs-Service] Successful close Elasticsearch")
}
func createIndexIfNotExists(client *elasticsearch.Client, index string) error {
	res, err := client.Indices.Exists([]string{index})
	if err != nil {
		log.Printf("[DEBUG] [Logs-Service] Error while checking Index is exist in Elasticsearch: %v", err)
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == 404 {
		mapping := `{
            "mappings": {
                "properties": {
                    "@timestamp": {"type": "date"},
                    "level":      {"type": "keyword"},
                    "message":    {"type": "text"}
                }
            }
        }`

		createRes, err := client.Indices.Create(
			index,
			client.Indices.Create.WithBody(strings.NewReader(mapping)),
		)
		if err != nil {
			log.Printf("[DEBUG] [Logs-Service] Error while created Index in Elasticsearch: %v", err)
			return err
		}
		defer createRes.Body.Close()
		if createRes.IsError() {
			log.Printf("[DEBUG] [Logs-Service] Failed to create Index in Elasticsearch: %v", createRes.String())
		}
	}
	return nil
}
