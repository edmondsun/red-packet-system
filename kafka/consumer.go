package kafka

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"red-packet-system/db"
	"red-packet-system/model"
	"red-packet-system/pkg/logger"
)

const consumerTimeout = 5 * time.Second // Set the maximum timeout for message processing

// StartConsumer starts the Kafka Consumer (supports multi-partitions)
func StartConsumer() {
	log := logger.GetLogger()

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	consumer, err := sarama.NewConsumer([]string{"kafka:9092"}, config)
	if err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}
	defer consumer.Close()

	partitions, err := consumer.Partitions("red_packet_transactions")
	if err != nil {
		log.Fatalf("Failed to get Kafka partitions: %v", err)
	}

	var wg sync.WaitGroup
	for _, partition := range partitions {
		wg.Add(1)
		go func(p int32) {
			defer wg.Done()
			consumePartition(consumer, p)
		}(partition)
	}

	wg.Wait()
}

// consumePartition consumes Kafka Partition
func consumePartition(consumer sarama.Consumer, partition int32) {
	pc, _ := consumer.ConsumePartition("red_packet_transactions", partition, sarama.OffsetNewest)
	defer pc.Close()

	for msg := range pc.Messages() {
		processKafkaMessage(msg)
	}
}

// processKafkaMessage processes Kafka message
func processKafkaMessage(msg *sarama.ConsumerMessage) {
	log := logger.GetLogger()
	log.Printf("Kafka message received: %s", string(msg.Value))

	data := strings.Split(string(msg.Value), ",")
	if len(data) != 3 {
		log.Println("Kafka message format error")
		log.Printf("Exiting processKafkaMessage, message format error")
		return
	}

	userID, err1 := strconv.ParseUint(data[0], 10, 64)
	log.Printf("Message parsing: userID=%d, err1=%v", userID, err1)
	redPacketID, err2 := strconv.ParseUint(data[1], 10, 64)
	log.Printf("Message parsing: redPacketID=%d, err2=%v", redPacketID, err2)
	amount, err3 := strconv.ParseFloat(data[2], 64)
	log.Printf("Message parsing: amount=%.2f, err3=%v", amount, err3)
	if err1 != nil || err2 != nil || err3 != nil {
		log.Println("Kafka message parsing error")
		log.Printf("Exiting processKafkaMessage, message parsing error")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), consumerTimeout)
	defer cancel()

	err := retryWithBackoff(ctx, func() error {
		log.Printf("retryWithBackoff: Calling updateUserBalance, userID=%d, redPacketID=%d, amount=%.2f", userID, redPacketID, amount)
		updateErr := updateUserBalance(uint(userID), uint(redPacketID), amount)
		log.Printf("retryWithBackoff: updateUserBalance returned, err=%v", updateErr)
		return updateErr
	}, 3)

	if err != nil {
		log.Printf("Kafka consumption failed: %v", err)
	} else {
		log.Printf("User %d grabbed %.2f Yuan (RedPacket %d)", userID, amount, redPacketID)
	}
	log.Printf("Exiting processKafkaMessage, message processing completed")
}

// updateUserBalance updates user balance
func updateUserBalance(userID uint, redPacketID uint, amount float64) error {
	dbInstance := db.GetDB()
	log := logger.GetLogger()

	log.Printf("Entering updateUserBalance, userID=%d, redPacketID=%d, amount=%.2f", userID, redPacketID, amount)

	if dbInstance == nil {
		err := fmt.Errorf("Database connection not initialized")
		log.Printf("Exiting updateUserBalance, database connection not initialized, err=%v", err)
		return err
	}

	var user model.User
	log.Printf("updateUserBalance: Executing dbInstance.First(&user, userID), userID=%d", userID)
	if err := dbInstance.First(&user, userID).Error; err != nil {
		log.Printf("updateUserBalance: dbInstance.First(&user, userID) failed, userID=%d, err=%v", userID, err, err)
		err = fmt.Errorf("User %d not found, err=%v", userID, err)
		log.Printf("Exiting updateUserBalance, User not found, err=%v", err)
		return err
	}
	log.Printf("updateUserBalance: dbInstance.First(&user, userID) succeeded, found User: %+v", user)

	user.Balance += amount
	log.Printf("updateUserBalance: Preparing to update User Balance, userID=%d, newBalance=%.2f", userID, user.Balance)
	err := dbInstance.Model(&user).Update("Balance", user.Balance).Error
	if err != nil {
		log.Printf("updateUserBalance: dbInstance.Model(&user).Update() failed, userID=%d, err=%v", userID, err, err)
		log.Printf("Exiting updateUserBalance, update failed, err=%v", err)
		return err
	}

	log.Printf("User %d balance updated: %.2f Yuan", userID, user.Balance)
	log.Printf("Exiting updateUserBalance, update successful")
	return nil
}
