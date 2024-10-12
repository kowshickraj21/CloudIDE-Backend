package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

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

func createStash(path string) string {
	buildCmd := exec.Command("docker","build",path)
	image, err := buildCmd.CombinedOutput()
	if err != nil{
		fmt.Println(err);
	}

	runCmd := exec.Command("docker","run",string(image))
	container, err := runCmd.CombinedOutput()
	if err != nil{
		fmt.Println(err);
	}
	return string(container)
}

func startHandler(w http.ResponseWriter, r *http.Request){
	var body string
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil{
		fmt.Println(err);
	}
	
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

func main(){
	http.HandleFunc("/ws",startHandler)
	fmt.Println("Server Listening at Port:3000")
	http.ListenAndServe(":5000",nil);
}