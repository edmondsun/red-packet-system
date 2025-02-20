package service

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math/rand"
	"time"

	"red-packet-system/db"
	"red-packet-system/kafka"
	"red-packet-system/model"
	"red-packet-system/pkg/logger"
	"red-packet-system/redisclient"

	"github.com/redis/go-redis/v9"
)

// Lua script for atomic stock decrement in Redis.
var luaScript = redis.NewScript(`
    local stock = redis.call("GET", KEYS[1])
    if not stock then
        return -2 -- No data in Redis, fallback to MySQL
    end
    if tonumber(stock) <= 0 then
        return -1 -- No more red packets available
    else
        redis.call("DECR", KEYS[1])
        return tonumber(stock) - 1
    end
`)

// GrabRedPacket handles red packet grabbing logic.
func GrabRedPacket(userID uint, redPacketID uint) (float64, error) {
	log := logger.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbInstance := db.GetDB()
	redisClient := redisclient.GetRedisClient()
	redlock := redisclient.GetRedlock()
	redisKey := fmt.Sprintf("red_packet_%d", redPacketID)
	lockKey := fmt.Sprintf("lock:red_packet_%d", redPacketID)

	// Check Bloom Filter before querying MySQL to prevent cache penetration
	if !redisclient.ExistsInBloomFilter(redisClient, redPacketID) {
		log.Println("[INFO] Red packet ID not found in Bloom Filter, rejecting request")
		return 0, errors.New("red packet does not exist")
	}

	// Acquire Redlock (Minimizing lock duration)
	mutex := redlock.NewMutex(lockKey)
	if err := mutex.Lock(); err != nil {
		log.Println("[ERROR] Failed to acquire Redis lock:", err)
		return 0, errors.New("system is busy, please try again later")
	}
	defer mutex.Unlock()

	// Execute Lua script for atomic stock decrement in Redis
	result, err := luaScript.Run(ctx, redisClient, []string{redisKey}).Int()
	if err != nil && err != redis.Nil {
		log.Println("[ERROR] Redis operation failed:", err)
		return 0, errors.New("system error")
	}

	// No red packets left
	if result == -1 {
		log.Println("[INFO] Red packet is already empty")
		return 0, errors.New("red packet is empty")
	}

	// Redis cache miss, check MySQL
	if result == -2 {
		var redPacket model.RedPacket
		if err := dbInstance.WithContext(ctx).First(&redPacket, redPacketID).Error; err != nil {
			log.Println("[ERROR] Red packet does not exist, rolling back Redis operation")
			redisClient.Set(ctx, redisKey, 0, 0)
			return 0, errors.New("red packet does not exist")
		}

		// 🌟 Set Redis cache with randomized TTL to prevent cache avalanche
		ttl := time.Duration(600+rand.Intn(60)) * time.Second
		redisClient.Set(ctx, redisKey, redPacket.RemainingCount, ttl)
		result = redPacket.RemainingCount
	}

	// Ensure red packet stock is available
	if result <= 0 {
		log.Println("[INFO] Red packet is already empty")
		return 0, errors.New("red packet is empty")
	}

	// Retrieve red packet data from MySQL
	var redPacket model.RedPacket
	if err := dbInstance.WithContext(ctx).First(&redPacket, redPacketID).Error; err != nil {
		log.Println("[ERROR] Red packet does not exist, rolling back Redis")
		redisClient.Incr(ctx, redisKey)
		return 0, errors.New("red packet does not exist")
	}

	// **Calculate amount to grab**
	amount := redPacket.RemainingAmount / float64(redPacket.RemainingCount)

	// **Use MySQL transaction to ensure consistency**
	err = dbInstance.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&redPacket).Updates(
			map[string]interface{}{
				"RemainingAmount": redPacket.RemainingAmount - amount,
				"RemainingCount":  redPacket.RemainingCount - 1,
			},
		).Error; err != nil {
			log.Println("[ERROR] Red packet update failed, rolling back Redis")
			redisClient.Incr(ctx, redisKey)
			return errors.New("red packet update failed")
		}

		// **Log red packet transaction**
		logEntry := model.RedPacketLog{
			UserID:      userID,
			RedPacketID: redPacketID,
			Amount:      amount,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := tx.Create(&logEntry).Error; err != nil {
			log.Println("[ERROR] Failed to log red packet grab, rolling back Redis")
			redisClient.Incr(ctx, redisKey)
			return errors.New("failed to log red packet grab")
		}

		return nil
	})

	if err != nil {
		log.Println("[ERROR] Transaction failed, rolling back Redis")
		redisClient.Incr(ctx, redisKey)
		return 0, err
	}

	// **Send Kafka event asynchronously**
	go kafka.SendToKafka(int64(userID), int64(redPacketID), amount)

	log.Printf("[SUCCESS] User %d grabbed %.2f from Red Packet %d\n", userID, amount, redPacketID)
	return amount, nil
}
