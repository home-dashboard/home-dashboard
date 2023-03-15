package cache

type Cache struct {
	Name  string
	items map[string]any
}

func (c *Cache) Set(key string, value any) {
	c.items[key] = value
}

func (c *Cache) Get(key string) (any, bool) {
	value, ok := c.items[key]
	return value, ok
}

func New(name string) *Cache {
	return &Cache{
		Name:  name,
		items: map[string]any{},
	}
}
