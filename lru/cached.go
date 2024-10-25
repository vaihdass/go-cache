package lru

import "time"

type cached[K comparable, V any] struct {
	key       K
	value     V
	expiredAt time.Time
}

func (c *cached[K, V]) expired(now time.Time) bool {
	return c.expiredAt.Before(now)
}
