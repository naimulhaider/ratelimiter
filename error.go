package ratelimiter

import "errors"

var MaxCapacityError = errors.New("exceeded max allocated tokens")