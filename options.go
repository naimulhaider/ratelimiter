package ratelimiter

import "github.com/go-redis/redis"

type Option func (*Limiter) error

func WithRedisStorage(client *redis.Client) Option {
	return func(l *Limiter) error {
		l.store = NewRedisStorage(client, l.maxRequests, l.interval)
		return nil
	}
}

func WithHTTPOptions(httpOptions HTTPOptions) Option {
	return func(l *Limiter) error {
		l.httpOptions = httpOptions
		return nil
	}
}