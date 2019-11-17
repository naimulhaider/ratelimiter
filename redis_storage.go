package ratelimiter

import (
	"errors"
	"github.com/go-redis/redis"
	"time"
)

const RetriesCount = 100

type RedisStorage struct {
	client *redis.Client
	maxTokens int64
	interval time.Duration
}

func NewRedisStorage(client *redis.Client, maxTokens int64, interval time.Duration) Storage {
	return &RedisStorage{
		client: client,
		maxTokens: maxTokens,
		interval: interval,
	}
}

func (s RedisStorage) UseToken(key string) (bool, error) {

	txFunc := func(tx *redis.Tx) error {
		currentTokens, err := tx.Get(key).Int64()
		if err != nil {
			if err == redis.Nil {
				// key doesn't exist, set tokens = 1 and expire the key at interval
				_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
					pipe.Set(key, 1, s.interval) // pipe handles the error case
					return nil
				})
			}
			return err
		}

		if currentTokens >= s.maxTokens {
			return MaxCapacityError // max capacity reached
		}

		// using a token, increment count
		_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
			pipe.Incr(key) // pipe handles the error case
			return nil
		})

		return err
	}

	// in case some other concurrent process completed before us, watch will fail and we will retry
	for retries := RetriesCount; retries > 0; retries-- {
		err := s.client.Watch(txFunc, key)
		if err != nil {
			if err == redis.TxFailedErr {
				continue // retry
			}

			if err == MaxCapacityError {
				return false, nil // max capacity reached, not an error
			}

			return false, err // some other error
		}

		return true, nil
	}

	return false, errors.New("max number of retries reached in UseToken")
}