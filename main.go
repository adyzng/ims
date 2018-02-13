package main

import (
	"flag"
	"ims/chat"
	"net/http"
	_ "net/http/pprof"

	"github.com/go-clog/clog"
	"github.com/gorilla/mux"
)

func init() {
	clog.New(clog.CONSOLE, clog.ConsoleConfig{
		Level:      clog.TRACE,
		BufferSize: 1024,
	})
}

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8080", "http service address")
	flag.Parse()

	hub := chat.NewChatHub()
	hub.NewRoom("test")

	hub.AddHandlers(
		func(msg *chat.Message, hub *chat.RoomHub) {
			if msg.Type == chat.T_MESSAGE {
				clog.Trace("new message %+v", msg)
			} else {
				clog.Info("cmd: %s, from: %s", msg.Type, msg.From)
			}
		},
	)

	r := mux.NewRouter()
	r.HandleFunc("/", IndexHandle)
	r.HandleFunc("/ws", hub.ServeWebsocket)
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	//http.HandleFunc("/", serveHome)
	//http.HandleFunc("/ws", hub.ServeWebsocket)

	defer clog.Shutdown()
	clog.Info("%v", http.ListenAndServe(addr, r))
}
