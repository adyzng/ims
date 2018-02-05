package chat

import (
	"net/http"

	"ims/std"
)

const (
	ch32 = 32
)

type Hub struct {
	broadcast std.Queue
	online    chan *Client
	offline   chan *Client
	clients   map[uint64]*Client
}

func NewHub() *Hub {
	h := &Hub{
		broadcast: std.NewSyncQueue(maxQueueSize),
		online:    make(chan *Client, ch32),
		offline:   make(chan *Client, ch32),
		clients:   make(map[uint64]*Client),
	}

	go h.run()
	return h
}

func (h *Hub) run() {
	chSend := h.broadcast.Cout()
	for {
		select {
		case c := <-h.online:
			h.clients[c.id] = c
			break

		case c := <-h.offline:
			if _, ok := h.clients[c.id]; ok {
				delete(h.clients, c.id)
				c.send.Close()
			}
			break

		case itm, ok := <-chSend:
			if !ok {
				return
			}
			msg, _ := itm.(*Message)
			for id, c := range h.clients {
				if !c.send.Add(msg) {
					delete(h.clients, id)
				}
			}
			break
		}
	}
}

func (h *Hub) ServeWsChat(w http.ResponseWriter, r *http.Request) {
	serveChartHandler(h, w, r)
}
