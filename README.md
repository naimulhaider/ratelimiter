# API Rate Limiter

#### Overview

Uses the [Token Bucket Algorithm](https://en.wikipedia.org/wiki/Token_bucket) to rate limit HTTP requests.
It allows only a specific number of requests to be processed in an interval, exceeding the configured rate will any decline further requests within the provided interval.


Safe to be used concurrently and in distributed networks. 
Only [Redis](https://redis.io/) is the supported storage for usage in distributed services.

Can be used directly or as a middleware for Go HTTP Handlers.
Response header and data can be configured when used as a middleware.

#### Usage
```
rlConfig := ratelimiter.Config {
    MaxRequests: 10,
    Interval: time.Minute,
}

redisStorage := ratelimiter.WithRedisStorage(redis.NewClient(
    &redis.Options {
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB    
    },
))

rl := ratelimiter.New(rlConfig, ratelimiter.WithRedisStorage(redisStorage))
```

###### Direct Usage

```
limiter := NewRateLimiter(10, time.Second, WithRedisStorage(redisClient))
success, err := limiter.IsAllowed(td.Key)
```

###### Middleware

```
limiter := NewRateLimiter(10, time.Second, WithRedisStorage(redisClient))
http.Handle("/helloworld", limiter.HTTPMiddleware(someHTTPHandler))
http.ListenAndServe(":8090", nil)
```

#### Internals

Storage is a key-value store that must be safe for concurrent operations. When local storage (Go Map) is used, a lock is used to control access to map keys. 

Redis operates on a single thread and instances of the services using this library should use the same redis database to allow distributed usage. 

The rate limiting logic works as follows:

###### With Redis Storage:

When `UseToken(key)` is called

A redis transaction is opened and operations are executed in a pipeline to ensure parallel changes by some other process is gracefully handled. The key expiry functionality is delegated to redis.
- If this key doesn't exist in storage, a new entry is created with tokens = 1 and expiry = interval. Function returns true.
- If current tokens exceeds max token limit, `MaxCapacityError` is sent. Function returns false.
- Tokens are incremented at the specified key. Function returns true.

###### With Local Storage:

When `UseToken(key)` is called

A lock is created so that concurrent access to the maps are restricted.
Two maps are created, one for tokens count and one for last refilled at timestamp. 

- If this key doesn't exist in storage, a new entry is created with tokens = 1 and lastRefilledAt = time.Now(). Function returns true.
- if lastRefilledAt + interval < currentTime, then its refilled as of previous step. Function returns true.
- If current tokens exceeds max token limit, Function returns false.
- Otherwise token is incremented, Function returns true.

```
TODO:
- Add support for concurrent access on specific keys instead of global lock
- Add functionality to clear unused keys at interval 
```