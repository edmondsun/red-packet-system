package kafka

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"red-packet-system/config"
)

const producerTimeout = 5 * time.Second // Define message timeout

// KafkaProducerSingleton ensures a single instance of Kafka producer
var (
	producer     sarama.SyncProducer
	producerOnce sync.Once
)

// initProducer initializes Kafka Sync Producer
func initProducer(cfg *config.Config) error {
	var err error
	producerOnce.Do(func() {
		saramaConfig := sarama.NewConfig()
		saramaConfig.Producer.Return.Successes = true
		saramaConfig.Producer.Retry.Max = 5

		producer, err = sarama.NewSyncProducer(cfg.KafkaBrokers, saramaConfig)
		if err != nil {
			log.Fatalf("Failed to initialize Kafka Producer: %v", err)
		} else {
			log.Println("Kafka Producer successfully initialized")
		}
	})

	return err
}

// SendToKafka sends a message to Kafka topic `red_packet_transactions`
func SendToKafka(userID, redPacketID int64, amount float64) error {
	cfg := config.LoadConfig()
	err := initProducer(cfg)
	if err != nil {
		return fmt.Errorf("Failed to initialize Kafka Producer: %v", err)
	}

	// Construct Kafka message
	message := &sarama.ProducerMessage{
		Topic:     "red_packet_transactions",
		Partition: int32(userID % 5), // Distribute across partitions
		Value:     sarama.StringEncoder(fmt.Sprintf("%d,%d,%.2f", userID, redPacketID, amount)),
	}

	ctx, cancel := context.WithTimeout(context.Background(), producerTimeout)
	defer cancel()

	// Retry mechanism
	err = retryWithBackoff(ctx, func() error {
		_, _, err := producer.SendMessage(message)
		return err
	}, 3)

	if err != nil {
		log.Printf("Failed to send Kafka message: %v", err)
		return err
	}

	log.Printf("Kafka message sent: UserID=%d, RedPacketID=%d, Amount=%.2f", userID, redPacketID, amount)
	return nil
}
