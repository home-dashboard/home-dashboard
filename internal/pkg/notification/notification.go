package notification

import (
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/teivah/broadcast"
)

var logger = comfy_log.New("[notification]")

type Message struct {
	Type string
	Data map[string]any
}

var relay *broadcast.Relay[Message]

func GetListener() *broadcast.Listener[Message] {
	return relay.Listener(1)
}

func Send(msgType string, data map[string]any) {
	relay.Broadcast(Message{Type: msgType, Data: data})
}

func init() {
	relay = broadcast.NewRelay[Message]()
}
