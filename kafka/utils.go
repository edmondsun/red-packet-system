package kafka

import (
	"context"
	"red-packet-system/pkg/logger"
	"time"
)

// maxKafkaRetries defines the maximum number of retries
const maxKafkaRetries = 3

// retryWithBackoff retries the operation with exponential backoff
func retryWithBackoff(ctx context.Context, operation func() error, maxKafkaRetries int) error {
	log := logger.GetLogger()

	var err error
	backoff := 100 * time.Millisecond // Initial backoff duration

	for i := 0; i < maxKafkaRetries; i++ {
		select {
		case <-ctx.Done():
			log.Println("Context timeout reached, stopping retries")
			return ctx.Err()
		default:
			err = operation()
			if err == nil {
				return nil // Operation succeeded
			}

			log.Printf("Operation failed (attempt %d/%d): %v", i+1, maxKafkaRetries, err)
			time.Sleep(backoff)
			backoff *= 2 // Double the backoff duration for each retry
		}
	}
	return err
}
