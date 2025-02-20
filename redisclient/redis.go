package redisclient

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"red-packet-system/config"
)

const maxRedisRetries = 5

var ctx = context.Background()
var redisClient *redis.ClusterClient

// InitRedis initializes Redis Cluster with retry mechanism.
func InitRedis(cfg *config.Config) *redis.ClusterClient {
	opts := &redis.ClusterOptions{
		Addrs:    cfg.RedisCluster,
		Password: cfg.RedisPassword,
	}

	var err error
	for i := 0; i < maxRedisRetries; i++ {
		redisClient = redis.NewClusterClient(opts)
		_, err = redisClient.Ping(ctx).Result()
		if err == nil {
			log.Println("[INFO] Redis Cluster connected successfully")
			return redisClient
		}

		log.Printf("[WARN] Redis connection attempt %d failed: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	log.Fatalf("[ERROR] Redis Cluster connection failed after retries: %v", err)
	return nil
}

// GetRedisClient returns the initialized Redis Cluster client.
func GetRedisClient() *redis.ClusterClient {
	if redisClient == nil {
		log.Fatal("[ERROR] Redis Cluster is not initialized")
	}
	return redisClient
}

// GetRedlock initializes a Redlock distributed lock for concurrency control.
func GetRedlock() *redsync.Redsync {
	if redisClient == nil {
		log.Fatal("[ERROR] Redis Cluster is not initialized, cannot create Redlock")
	}

	pool := goredis.NewPool(redisClient)
	return redsync.New(pool)
}

// ExistsInBloomFilter checks if the red packet ID exists in the Bloom Filter to prevent cache penetration.
func ExistsInBloomFilter(client *redis.ClusterClient, redPacketID uint) bool {
	bloomKey := "bloom_filter:red_packets"
	exists, err := client.Exists(ctx, bloomKey, strconv.FormatUint(uint64(redPacketID), 10)).Result()
	if err != nil {
		log.Printf("[WARN] Error checking Bloom Filter: %v", err)
		return true // Assume it exists to prevent unnecessary DB queries.
	}
	return exists > 0
}
