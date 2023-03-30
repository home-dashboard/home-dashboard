package sessions

import (
	"github.com/gin-contrib/sessions"
	gormSessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/database"
)

var sessionName = "notificationSession"

var store *gormSessions.Store

func initialStore() *gormSessions.Store {
	store := gormSessions.NewStore(database.GetDB(), true, []byte("secret"))
	return &store
}

//func GetStore() *gormSessions.Store {
//	return store
//}

var middleware gin.HandlerFunc

// GetSessionMiddleware 获取自定义的 session 中间件
func GetSessionMiddleware() gin.HandlerFunc {
	if store == nil {
		store = initialStore()
	}

	if middleware == nil {
		middleware = sessions.Sessions(sessionName, *store)
	}

	return middleware
}
