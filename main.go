package main

import (
	"flag"
	"fmt"
	"ims/chat"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime/debug"

	"github.com/go-clog/clog"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func init() {
	clog.New(clog.CONSOLE, clog.ConsoleConfig{
		Level:      clog.TRACE,
		BufferSize: 1024,
	})
}

func main() {
	var addr string

	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			fmt.Printf("panic: %v", err)
		}
	}()

	flag.StringVar(&addr, "addr", ":8080", "http service address")
	flag.Parse()

	hub := chat.NewHub()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", hub.ServeWsChat)

	defer clog.Shutdown()
	clog.Info("%v", http.ListenAndServe(addr, nil))
}
