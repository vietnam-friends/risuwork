package scenario

import (
	"fmt"
	"math/rand"
	"sync"
)

type Container[T comparable] struct {
	mutex sync.RWMutex
	ts    []*T
	tMap  map[T]any
}

func (c *Container[T]) Random() (*T, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if len(c.ts) == 0 {
		return nil, false
	}
	return c.ts[rand.Intn(len(c.ts))], true
}

func (c *Container[T]) Add(t T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.ts = append(c.ts, &t)
	if c.tMap == nil {
		c.tMap = make(map[T]any)
	}
	c.tMap[t] = struct{}{}
}

func (c *Container[T]) Exists(t T) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.tMap == nil {
		return false
	}
	_, ok := c.tMap[t]
	return ok
}

func (c *Container[T]) GoString() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return fmt.Sprintf("type: %T, len: %d", c, len(c.ts))
}
