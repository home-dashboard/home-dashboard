package cache_control

import (
	"github.com/gin-gonic/gin"
)

// CacheControlMiddleware Cache-Control 中间件
// value 将被添加到响应头中, value 可以是任何合法的 Cache-Control 值, 例如: "no-cache, no-store, must-revalidate".
//
// # 对于静态文件的最佳实践
//
//	Cache-Control: public, max-age=604800, immutable
//
// 对于静态资源，现代的最佳实践是在其URL中包含版本/哈希，同时永不修改资源 - 但是，必要时，用新的版本更新资源，这些新的版本有新的版本号/哈希，使得它们的URL不同。这被称为破坏缓存模式。
//
//	<script src=https://example.com/react.0.0.0.js></script>
//
// [Cache-Control]: https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Cache-Control
func CacheControlMiddleware(value string) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Header("Cache-Control", value)
		context.Next()
	}
}
