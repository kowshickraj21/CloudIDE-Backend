package main

import (
	"log"
	"main/ws"
	"net/http"
)

func main() {

	client := AWSInit()
	if client == nil {
		log.Fatalln("Initialization Error!")
	}
	
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.StartSocket(w,r,client)
	})
}
