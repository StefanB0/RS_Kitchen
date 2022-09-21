package pkg

import "sync"

type Counter struct {
	I  int
	mu sync.Mutex
}

func (c *Counter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.I++
}

func (c *Counter) Deincrement() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.I--
}

func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.I
}

func removeDish(s []KitchenDish, i int) []KitchenDish {
	return append(s[:i], s[i+1:]...)
}