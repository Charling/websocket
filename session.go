package ws

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	LOGGER "github.yn.com/ext/common/logger"
	VK_Proto "github.yn.com/ext/common/proto"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
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

func (c *Session) GetID() int64 {
	return c.id
}

// 关闭连接
func (c *Session) Close() {
	c.close()
}

func (c *Session) close() {
	c.hub.unregister <- c
	c.conn.Close()
}

func (c *Session) ResetWaitTime() {
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
}

func (c *Session) read() {
	defer c.close()

	c.conn.SetReadLimit(maxMessageSize)

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

			/*	case <-ticker.C:
				c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			*/
		}
	}
}

func (c *Session) SendDataToClient(ops int32, playerId int64, data []byte) bool {
	msg := &VK_Proto.Message{
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

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	mutex.Lock()
	serial = serial + 1
	id := serial
	mutex.Unlock()
	fmt.Printf("a new ws conn: %s->%s sid:%d\n", conn.RemoteAddr().String(), conn.LocalAddr().String(), id)
	s := &Session{hub: hub, conn: conn, send: make(chan []byte, 256), id: id}
	s.hub.register <- s

	go s.write()
	go s.read()
}
