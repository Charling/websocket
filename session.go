package ws

import (
	"log"
	"net"
	"sync"
	"time"

	LOGGER "github.yn.com/ext/common/logger"
	Proto "github.yn.com/ext/common/proto"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

var (
	newline  = []byte{'\n'}
	space    = []byte{' '}
	upgrader = websocket.Upgrader{
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
	}
	mutex sync.Mutex
)

type Session struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	id   int64
}

func (c *Session) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Session) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Session) GetId() int64 {
	return c.id
}

// 关闭连接
func (c *Session) Close() {
	c.hub.unregister <- c
	c.conn.Close()
}

func (c *Session) ResetWaitTime() {
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })
}

func (c *Session) read() {
	defer c.Close()

	c.conn.SetReadLimit(1024)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		onMessage(c, message)
	}
}

func (c *Session) write() {
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.BinaryMessage, message)
		}
	}
}

func (c *Session) SendDataToClient(ops int32, playerId int64, data []byte) bool {
	msg := &Proto.Message{
		PlayerId: proto.Int64(playerId),
		Ops:      proto.Int32(ops),
		Data:     data,
	}

	final, err := proto.Marshal(msg)
	if err != nil {
		LOGGER.Error("SendDataToClient SerializeFailed")
		return false
	}

	c.send <- final
	return true
}
