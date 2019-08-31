package ws

import(
    "net/http"
    LOGGER "github.yn.com/ext/common/logger"
)

var (
    hub *Hub
    serial int64
)

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