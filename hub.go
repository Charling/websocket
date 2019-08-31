
package ws

type Hub struct {
	sessions map[int64] *Session
	register chan *Session
	unregister chan *Session
}

func newHub() *Hub {
	return &Hub{
		register: make(chan *Session),
		unregister: make(chan *Session),
		sessions: make(map[int64] *Session),
	}
}

func (h *Hub) run() {
	for {
		select {
		case s := <-h.register:
			h.sessions[s.id] = s
			onOpen(s)
			
		case s := <-h.unregister:
			if _, ok := h.sessions[s.id]; ok {
				onClose(s)
				delete(h.sessions, s.id)
				close(s.send)
			} 
		}
	}
}
