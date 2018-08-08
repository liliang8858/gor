package main

import (
	"net/http"
	"./controller"
		"github.com/alfred-zhong/wserver"
	"log"
)


func main() {
	// Run server websocket
	go startWebsocket()

	// Run server web
	startWeb()
}

// web
func startWeb() {
	log.Println("Listening...")
	http.HandleFunc("/", controller.Dispatcher)
	http.ListenAndServe(":3000", nil)
}

// websocket
func startWebsocket(){
	log.Println("Listening ws start...")
	// Run server
	server := wserver.NewServer(":3001")
	server.WSPath = "/ws"
	server.PushPath = "/push"
	server.AuthToken = func(token string) (userID string, ok bool) {
		if token == "aaa" {
			return "jack", true
		}
		return "", false
	}
	server.PushAuth = func(r *http.Request) bool {
		return true
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}

	log.Println("Listening ws ...")

}
