package notification

import (
	"github.com/teivah/broadcast"
)

type Message struct {
	Type string
	Data map[string]any
}

var relay *broadcast.Relay[Message]

func GetListener() *broadcast.Listener[Message] {
	return relay.Listener(1)
}

func Send(_type string, data map[string]any) {
	relay.Broadcast(Message{Type: _type, Data: data})
}

func init() {
	relay = broadcast.NewRelay[Message]()
}
