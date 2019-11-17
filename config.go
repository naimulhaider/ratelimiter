package ratelimiter

import "net/http"

const (
	REDIS_TEST_DB = 2
	HTTPRateLimitingStatusCode = http.StatusTooManyRequests
	HTTPJSONContentType = "application/json"
	DefaultHeaderIPKey = "RemoteAddr"
	DefaultRateLimitedResponseMessage = "Rate Limited! Too many requests."
)

type HTTPOptions struct {
	HeaderIPKey string // key from which ip address is read
	ResponseStatusCode int // status code of rate limited repsonse
	ResponseContentType string // content type of rate limited response
	ResponseMessage string // the message to respond with in case of failure
}

func DefaultHTTPOptions() HTTPOptions {
	return HTTPOptions {
		HeaderIPKey: DefaultHeaderIPKey,
		ResponseStatusCode: HTTPRateLimitingStatusCode,
		ResponseContentType: HTTPJSONContentType,
		ResponseMessage: DefaultRateLimitedResponseMessage,
	}
}