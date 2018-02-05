package chat

import (
	"net/http"
	"strconv"
	"time"

	"ims/std"

	"github.com/go-clog/clog"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// Maximum send queue size
	maxQueueSize = 1024
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	id   uint64
	hub  *Hub
	conn *websocket.Conn
	send std.Queue
}

func (c *Client) readPump() {
	defer func() {
		c.hub.offline <- c
		c.send.Close()
		c.conn.Close()
		clog.Info("Client %d read routine end.", c.id)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				clog.Error(2, "client %v read message error: %v", c.id, err)
			} else {
				clog.Trace("client %v closed.", c.id)
			}
			return
		}

		msg.ID = c.id
		msg.From = strconv.FormatUint(c.id, 10)
		msg.Timestamp = std.GetNowMs()

		c.hub.broadcast.Add(&msg)
		clog.Trace("receive client %v message: %v.", c.id, msg.Content)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.send.Close()
		c.conn.Close()
		clog.Info("client %d write routine end.", c.id)
	}()

	var err error
	var msg *Message
	chMsg := c.send.Cout()

	for {
		select {
		case v, ok := <-chMsg:
			if !ok && c.hub.broadcast.Closed() {
				clog.Warn("websocket server closed.")
				return
			}
			if msg, ok = v.(*Message); !ok {
				continue
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err = c.conn.WriteJSON(msg)

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					clog.Error(2, "client %d writer error: %v", c.id, err)
				} else {
					clog.Trace("client %v closed.", c.id)
				}
				return
			}
			clog.Trace("send message %v to client %v.", msg.Content, c.id)

			/*
				w, err := c.conn.NextWriter(websocket.TextMessage)

				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
						clog.Error(2, "client %d writer error: %v", c.id, err)
					} else {
						clog.Trace("client %v closed.", c.id)
					}
					return
				}

				c.conn.WriteJSON

				msg, _ := item.(*Message)
				data, _ := json.Marshal(msg)
				w.Write(data)

				if err := w.Close(); err != nil {
					clog.Error(2, "client %v write message error: %v", c.id, err)
					return
				}
				clog.Trace("send message %v to client %v.", msg.Content, c.id)
			*/

			break

		case <-ticker.C:
			// heartbeat with client
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err = c.conn.WriteMessage(websocket.PingMessage, nil)

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					clog.Error(2, "client %d heartbeat error: %v", c.id, err)
				} else {
					clog.Trace("client %v closed.", c.id)
				}
				return
			}
		}
	}
}

func serveChartHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		clog.Error(2, "websocket chat connection error: %v", err)
		return
	}

	c := &Client{
		id:   std.GenUniqueID(),
		hub:  hub,
		conn: conn,
		send: std.NewSyncQueue(maxQueueSize),
	}

	hub.online <- c
	clog.Info("New websocket client %d", c.id)

	// handler websocket read/write client in seperate goroutine
	go c.readPump()
	go c.writePump()
}
