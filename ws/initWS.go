package ws

import (
	"context"
	"fmt"
	"main/aws"
	"main/k8s"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
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

func StartSocket(w http.ResponseWriter,r *http.Request, client *s3.Client){

	bucket := os.Getenv("AWS_BUCKET")
	k8s.StartPod("check","nodejs")

	conn,err := upgrader.Upgrade(w,r,nil)
	if err != nil {
		fmt.Println(err);
	}
	fmt.Println("New Client:",conn.LocalAddr())
	defer conn.Close();

	for {
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			fmt.Println(err);
			break
		}
		switch message.Type {
		case "createObject":
			aws.CreateObject(context.TODO(),client,bucket, message.Data)
		case "deleteFolder":
			aws.DeleteS3Folder(context.TODO(),client,bucket,message.Data)
		case "getDir":
			objects,_ := aws.ListDirectory(context.TODO(),client,bucket,message.Data)
			conn.WriteJSON(objects)
		default:
			fmt.Println("Wrong Request occured!")
			conn.Close()
		} 
	}
	fmt.Println("Client Disconnected:", conn.LocalAddr())
}
