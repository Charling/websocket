package ws

import (
	"log"
	Proto "github.yn.com/ext/common/proto"
	"github.com/golang/protobuf/proto"
)

type Handler func(s *Session, msg *Proto.Message) 
type Interface func(s *Session)

const (
	Connect = 0
	DisConnect = 1
)

var (
	mapHandlers map[int32] Handler
)

func Register(msgHand *map[int32]Handler) {
	mapHandlers = *msgHand
}

func onOpen(s *Session) {
	if _, ok := mapHandlers[Connect]; ok {
		mapHandlers[Connect](s, nil)
	}
}

func onClose(s *Session) {
	if _, ok := mapHandlers[DisConnect]; ok {
		mapHandlers[DisConnect](s, nil)
	}
}

func onMessage(s *Session, data []byte) {
	msg := &Proto.Message { }
	err := proto.Unmarshal(data, msg)
	if err != nil {
		log.Println(err)
		return
	}

	h, exists := mapHandlers[*msg.Ops]
	if exists {
		h(s, msg)
	}
}
