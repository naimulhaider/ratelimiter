package ratelimiter

import (
	"sync"
	"time"
)

type LocalStorage struct {
	tokenMap map[string] int64
	lastRefilledAtMap map[string]int64
	maxTokens int64
	interval time.Duration
	lock *sync.Mutex
}

func NewLocalStorage(maxTokens int64, interval time.Duration) Storage {
	return &LocalStorage{
		tokenMap: make(map[string]int64),
		lastRefilledAtMap: make(map[string]int64),
		maxTokens: maxTokens,
		interval: interval,
		lock: &sync.Mutex{},
	}
}

func (s LocalStorage) UseToken(key string) (bool, error) {
	s.lock.Lock()

	defer s.lock.Unlock()

	curTmNano := time.Now().UnixNano()

	currentTokens := s.tokenMap[key]

	if currentTokens == 0 {
		s.tokenMap[key] = 1
		s.lastRefilledAtMap[key] = curTmNano
		return true, nil
	}

	if s.lastRefilledAtMap[key] + s.interval.Nanoseconds() < curTmNano {
		// needs refill
		s.tokenMap[key] = 1
		s.lastRefilledAtMap[key] = curTmNano
		return true, nil
	}

	if currentTokens >= s.maxTokens {
		return false, nil // max capacity reached
	}

	// using a token, increment token count
	s.tokenMap[key] = currentTokens + 1

	return true, nil
}