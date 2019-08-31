package ws

import(
    "net/http"
    LOGGER "github.yn.com/ext/common/logger"
    "log"
    "fmt"
)

var (
    hub *Hub
    serial int64
)

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

func GetSession(id int64) *Session {
    if _, ok := hub.sessions[id]; ok {
        return hub.sessions[id]
    }
    return nil
}

func OnStartup(addr string) {
    hub = newHub()
    go hub.run()
    
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
    })
    defer func() {
		err := http.ListenAndServe(addr, nil)
        if err != nil {
            LOGGER.Error("ListenAndServe: %s", err)
        }
	}()
}