package ratelimiter

import (
	"net/http"
	"time"
)

type Limiter struct {
	store Storage
	maxRequests int64
	interval time.Duration
	httpOptions HTTPOptions
}

func NewRateLimiter(maxRequests int64, interval time.Duration, opts ...Option) *Limiter {
	limiter := &Limiter{
		maxRequests: maxRequests,
		interval: interval,
		httpOptions: DefaultHTTPOptions(),
	}

	for _, op := range opts {
		op(limiter)
	}

	return limiter
}

func (l Limiter) IsAllowed(key string) (bool, error) {
	return l.store.UseToken(key)
}

func (l Limiter) HTTPMiddleware(nextFunc func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipAddress := r.Header.Get(l.httpOptions.HeaderIPKey)
		isAllowed, err := l.IsAllowed(ipAddress)
		if err != nil {
			w.Header().Set("Content-Type", HTTPJSONContentType)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if !isAllowed {
			w.Header().Set("Content-Type", l.httpOptions.ResponseContentType)
			w.WriteHeader(l.httpOptions.ResponseStatusCode)
			w.Write([]byte(l.httpOptions.ResponseMessage))
			return
		}
		http.HandlerFunc(nextFunc).ServeHTTP(w, r)
	})
}