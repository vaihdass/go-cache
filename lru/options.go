package lru

import "time"

type cacheOptions struct {
	capacity int
	ttl      time.Duration
}

type Option func(*cacheOptions)

// WithCapacity ignoring negative and zero capacity values
func WithCapacity(capacity int) Option {
	return func(o *cacheOptions) {
		if capacity > 0 {
			o.capacity = capacity
		}
	}
}

// WithTTL ignoring negative TTL values, zero value means TTL is not used
func WithTTL(ttl time.Duration) Option {
	return func(o *cacheOptions) {
		if ttl >= 0 {
			o.ttl = ttl
		}
	}
}
