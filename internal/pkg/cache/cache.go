package cache

import cache2 "github.com/patrickmn/go-cache"

var memoryCache = cache2.New(cache2.NoExpiration, 0)

type Cache struct {
	*cache2.Cache
	name string
}

func (c *Cache) Set(key string, config any) {
	memoryCache.SetDefault(key, config)
}

func (c *Cache) Get(key string) (any, bool) {
	return memoryCache.Get(key)
}

func New(name string) *Cache {
	cache := cache2.New(cache2.NoExpiration, 0)

	return &Cache{
		cache,
		name,
	}
}
