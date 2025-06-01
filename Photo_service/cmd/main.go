package main

import (
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/configs"
)

func main() {
	config := configs.LoadConfig()
	kafkaProducer := kafka.NewKafkaProducer(config.Kafka)
}
