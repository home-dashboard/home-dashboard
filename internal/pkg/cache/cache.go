package cache

import "sync"

type Cache struct {
	Name  string
	items sync.Map
}

func (c *Cache) Set(key string, value any) {
	c.items.Store(key, value)
}

func (c *Cache) Get(key string) (any, bool) {
	value, ok := c.items.Load(key)
	return value, ok
}

func New(name string) *Cache {
	return &Cache{
		Name:  name,
		items: sync.Map{},
	}
}
