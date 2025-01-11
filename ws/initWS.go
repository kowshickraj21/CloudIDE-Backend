package ws

import (
	"context"
	"fmt"
	"main/aws"
	"main/k8s"
	"net/http"
	"os"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/websocket"
)

type Message struct{
	Type string `json:"type"`
	Data string `json:"data"`
	Path string `json:"path"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
        return true
    },
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,	
}

func StartSocket(w http.ResponseWriter,r *http.Request, client *s3.Client){
	deploymentName := r.URL.Query().Get("stash")
	fmt.Println("deployment Name",deploymentName)
	bucket := os.Getenv("AWS_BUCKET")
	conn,err := upgrader.Upgrade(w,r,nil)
	if err != nil {
		fmt.Println(err);
	}
	fmt.Println("New Client:",conn.LocalAddr())
	defer conn.Close();
	terminal := k8s.StartTerminal(deploymentName)


	for {
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			fmt.Println("ERR:",err);
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
		case "getFile":
			fileData,err := aws.GetFile(context.TODO(),client,bucket,message.Data)
			if err == nil {
				conn.WriteJSON(fileData)
			}
		case "writeFile":
			aws.WriteFile(context.TODO(),client,bucket, message.Path,message.Data)
		case "terminalCommand":
			if terminal == nil {
				fmt.Println("No active terminal session")
				break
			}
			_, err := terminal.Stdin.Write([]byte(message.Data + "\n"))
			if err != nil {
				fmt.Println("Error writing to terminal:", err)
			}
			go func() {
				buf := make([]byte, 1024)
				for {
					n, err := terminal.Stdout.Read(buf)
					if err != nil {
						fmt.Println("Error reading from terminal stdout:", err)
						break
					}
					if n > 0 {
						output := Message{
							Type: "output",
							Data: stripANSI(string(buf[:n])),
						}
						conn.WriteJSON(output)
						fmt.Println("Output:",stripANSI(string(buf[:n])))
					}
				}
			}()
		default:
			fmt.Println("Wrong Request occured!")
			terminal.Close()
			k8s.CloseStash(deploymentName)
			conn.Close()
		} 
	}
	fmt.Println("Client Disconnected:", conn.LocalAddr())
}


func stripANSI(input string) string {
    re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
    return re.ReplaceAllString(input, "")
}