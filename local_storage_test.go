package ratelimiter

import (
	"sync"
	"testing"
	"time"
)

func TestLocalStorage_UseToken(t *testing.T) {

	localStorage := NewLocalStorage(10, time.Second)

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

					success, err := localStorage.UseToken(td.Key)
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
}