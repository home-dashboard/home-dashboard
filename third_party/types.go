package third_party

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/third_party/internal"
)

type ThirdPartyEventImpl = internal.ThirdPartyEvent

// 对 internal.ThirdPartyEvent[string, any] 的实现.
type thirdPartyEventImpl struct {
	Type string
	Data any
}

func (e thirdPartyEventImpl) GetType() string {
	return e.Type
}
func (e thirdPartyEventImpl) GetData() any {
	return e.Data
}

// newThirdPartyEvent 创建一个新的基础第三方事件.
func newThirdPartyEvent(_type string, data any) thirdPartyEventImpl {
	return thirdPartyEventImpl{
		Type: _type,
		Data: data,
	}
}

// NewNotificationChannelConnectedEvent 创建一个新的通知通道已连接事件.
func NewNotificationChannelConnectedEvent(context *gin.Context) ThirdPartyEventImpl {
	return internal.NotificationChannelConnectedEvent{
		ThirdPartyEvent: newThirdPartyEvent(internal.NotificationChannelConnectedEventType, internal.NotificationChannelConnectedEventData{
			Context: context,
		}),
	}
}
