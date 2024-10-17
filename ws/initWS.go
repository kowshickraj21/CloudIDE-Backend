package ws

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct{
	Type string `json:"type"`
	Data string `json:"data"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
        return true
    },
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,	
}

func startSocket(w http.ResponseWriter, r *http.Request){
	conn,err := upgrader.Upgrade(w,r,nil)
	if err != nil {
		fmt.Println(err);
	}
	fmt.Println("New Client:",conn.LocalAddr())
	defer conn.Close();

	for {
		var message Message
		err := conn.ReadJSON(&message)
		fmt.Println(message)
		if err != nil {
			fmt.Println(err);
			break
		} 
		err = conn.WriteJSON(Message{Type: "Response",Data:"Recieved"});
		if err != nil {
			fmt.Println(err);
		} 
	}
	fmt.Println("Client Disconnected:", conn.LocalAddr())
}