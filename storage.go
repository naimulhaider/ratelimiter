package ratelimiter

type Storage interface{
	UseToken(string) (bool, error)
}