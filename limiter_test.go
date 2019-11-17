package ratelimiter

import (
	"github.com/go-redis/redis"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestLimiter_IsAllowed(t *testing.T) {

	redisClient := redis.NewClient(
		&redis.Options {
			Addr:     "localhost:6379",
			Password: "",
			DB:       REDIS_TEST_DB,
		},
	)

	limiter := NewRateLimiter(10, time.Second, WithRedisStorage(redisClient))

	data := []struct {
		Key string
		RequestsCount int
		RequestPauseBetween time.Duration
		SuccessCount int
	}{
		{
			Key: "10.11.12.1",
			RequestsCount: 10,
			RequestPauseBetween: 0,
			SuccessCount: 10,
		},
		{
			Key: "10.11.12.2",
			RequestsCount: 15,
			RequestPauseBetween: 0,
			SuccessCount: 10,
		},
		{
			Key: "10.11.12.3",
			RequestsCount: 5,
			RequestPauseBetween: 0,
			SuccessCount: 5,
		},
		{
			Key: "10.11.12.4",
			RequestsCount: 100,
			RequestPauseBetween: 0,
			SuccessCount: 10,
		},
		{
			Key: "10.11.12.5",
			RequestsCount: 11,
			RequestPauseBetween: 0,
			SuccessCount: 10,
		},
		{
			Key: "10.11.12.6",
			RequestsCount: 9,
			RequestPauseBetween: 0,
			SuccessCount: 9,
		},
		{
			Key: "10.11.12.7",
			RequestsCount: 20,
			RequestPauseBetween: 10 * time.Millisecond,
			SuccessCount: 10,
		},
		{
			Key: "10.11.12.8",
			RequestsCount: 10,
			RequestPauseBetween: 80 * time.Millisecond,
			SuccessCount: 10,
		},
		{
			Key: "10.11.12.9",
			RequestsCount: 18,
			RequestPauseBetween: 110 * time.Millisecond,
			SuccessCount: 18,
		},
	}

	dataWG := sync.WaitGroup{}

	for testIdx, testData := range data {
		ti := testIdx
		td := testData

		dataWG.Add(1)

		go func() {
			defer dataWG.Done()

			useWG := sync.WaitGroup{}
			successCount := 0
			counterMutex := sync.Mutex{}

			for i := 0; i < td.RequestsCount; i++ {
				useWG.Add(1)

				go func() {
					defer useWG.Done()

					success, err := limiter.IsAllowed(td.Key)
					if err != nil {
						t.Errorf("Didn't expect error, got: %v", err.Error())
						return
					}

					if success {
						counterMutex.Lock()
						successCount++
						counterMutex.Unlock()
					}

					return
				}()

				time.Sleep(td.RequestPauseBetween)
			}

			useWG.Wait()

			if successCount != td.SuccessCount {
				t.Errorf("Test: %d, Expected success count: %d, got success count: %d", ti, td.SuccessCount, successCount)
			}
		}()
	}

	dataWG.Wait()

	err := redisClient.FlushDB().Err()
	if err != nil {
		t.Errorf("Failed to flush redis db at end of test")
	}
}

func TestLimiter_HTTPMiddleware(t *testing.T) {

	helloWorldHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", HTTPJSONContentType)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`Hello World!`))
		return
	}

	redisClient := redis.NewClient(
		&redis.Options {
			Addr:     "localhost:6379",
			Password: "",
			DB:       REDIS_TEST_DB,
		},
	)

	limiter := NewRateLimiter(10, time.Second, WithRedisStorage(redisClient))

	data := []struct {
		IPAddr string
		RequestsCount int
		RequestPauseBetween time.Duration
		SuccessCount int
	}{
		{
			IPAddr: "10.11.12.1",
			RequestsCount: 10,
			RequestPauseBetween: 0,
			SuccessCount: 10,
		},
		{
			IPAddr: "10.11.12.2",
			RequestsCount: 15,
			RequestPauseBetween: 0,
			SuccessCount: 10,
		},
		{
			IPAddr: "10.11.12.3",
			RequestsCount: 5,
			RequestPauseBetween: 0,
			SuccessCount: 5,
		},
		{
			IPAddr: "10.11.12.4",
			RequestsCount: 100,
			RequestPauseBetween: 0,
			SuccessCount: 10,
		},
		{
			IPAddr: "10.11.12.5",
			RequestsCount: 11,
			RequestPauseBetween: 0,
			SuccessCount: 10,
		},
		{
			IPAddr: "10.11.12.6",
			RequestsCount: 9,
			RequestPauseBetween: 0,
			SuccessCount: 9,
		},
		{
			IPAddr: "10.11.12.7",
			RequestsCount: 20,
			RequestPauseBetween: 10 * time.Millisecond,
			SuccessCount: 10,
		},
		{
			IPAddr: "10.11.12.8",
			RequestsCount: 10,
			RequestPauseBetween: 80 * time.Millisecond,
			SuccessCount: 10,
		},
		{
			IPAddr: "10.11.12.9",
			RequestsCount: 18,
			RequestPauseBetween: 110 * time.Millisecond,
			SuccessCount: 18,
		},
	}

	dataWG := sync.WaitGroup{}

	for testIdx, testData := range data {
		ti := testIdx
		td := testData

		dataWG.Add(1)

		go func() {
			defer dataWG.Done()

			useWG := sync.WaitGroup{}
			successCount := 0
			counterMutex := sync.Mutex{}

			for i := 0; i < td.RequestsCount; i++ {
				useWG.Add(1)

				go func() {
					defer useWG.Done()

					req, err := http.NewRequest("GET", "/helloworld", nil)
					if err != nil {
						t.Fatal(err)
					}

					req.Header.Set(DefaultHeaderIPKey, td.IPAddr)

					rr := httptest.NewRecorder()
					handler := limiter.HTTPMiddleware(helloWorldHandler)
					handler.ServeHTTP(rr, req)

					if rr.Code == http.StatusOK {
						counterMutex.Lock()
						successCount++
						counterMutex.Unlock()
					}

					return
				}()

				time.Sleep(td.RequestPauseBetween)
			}

			useWG.Wait()

			if successCount != td.SuccessCount {
				t.Errorf("Test: %d, Expected success count: %d, got success count: %d", ti, td.SuccessCount, successCount)
			}
		}()
	}

	dataWG.Wait()

	err := redisClient.FlushDB().Err()
	if err != nil {
		t.Errorf("Failed to flush redis db at end of test")
	}
}