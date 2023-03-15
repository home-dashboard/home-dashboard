package cache

import "testing"

func BenchmarkCache_Get(b *testing.B) {
	type CachedObject struct {
		Type int
		Name string
	}

	cache := New("BenchmarkCache_Get")
	cache.Set("cachedObject", CachedObject{
		Type: 1,
		Name: "cachedObject",
	})

	for i := 0; i < 1e8; i++ {
		cachedValue, _ := cache.Get("cachedObject")
		_, _ = cachedValue.(CachedObject)

	}
}

func BenchmarkCache_Get_Point(b *testing.B) {
	type CachedObject struct {
		Type int
		Name string
	}

	cache := New("BenchmarkCache_Get_Point")
	cache.Set("cachedObject", &CachedObject{
		Type: 1,
		Name: "cachedObject",
	})

	for i := 0; i < 1e8; i++ {
		cache.Get("cachedObject")
	}
}
