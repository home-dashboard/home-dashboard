package sessions

import (
	"github.com/gin-contrib/sessions"
	gormSessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
)

var sessionName = "notificationSession"

var store = initialStore()

func initialStore() *gormSessions.Store {
	store := gormSessions.NewStore(monitor_db.GetDB(), true, []byte("secret"))
	return &store
}

//func GetStore() *gormSessions.Store {
//	return store
//}

var middleware gin.HandlerFunc

// GetSessionMiddleware 获取自定义的 session 中间件
func GetSessionMiddleware() gin.HandlerFunc {
	if middleware == nil {
		middleware = sessions.Sessions(sessionName, *store)
	}

	return middleware
}
