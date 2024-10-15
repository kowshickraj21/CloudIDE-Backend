package main

import (
	// "context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	// "os"
	"os/exec"

	// "github.com/aws/aws-sdk-go-v2/aws"
	// "github.com/aws/aws-sdk-go-v2/config"
	// "github.com/aws/aws-sdk-go-v2/credentials"
	// "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
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

func createStash(name string, image string) {
	userDir,_ := os.UserHomeDir()
	mount := fmt.Sprintf("%s:/app", filepath.Join("/home", userDir, "s3-bucket", "stashes", name))
	runCmd := exec.Command("docker", "run", "--name", name, "-v", mount, image)

	out, err := runCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error creating stash: %s\n", out)
		return
	}

	fmt.Println(string(out))
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

func main() {

	err := 	godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

	// staticCreds := aws.NewCredentialsCache(
	// 	credentials.NewStaticCredentialsProvider(
	// 		os.Getenv("AWS_ACCESS_KEY"),
	// 		os.Getenv("AWS_ACCESS_SECRET"), ""))
			
	// cfg, err := config.LoadDefaultConfig(
	// 	context.TODO(),
	// 	config.WithRegion(os.Getenv("AWS_ACCESS_REGION")),
	// 	config.WithCredentialsProvider(staticCreds),
	// )
	// if err != nil {
	// 	log.Fatalf("unable to load SDK config, %v", err)
	// }

	// s3Client := s3.NewFromConfig(cfg)

	// bucket := os.Getenv("AWS_BUCKET")
	// dstPrefix := "new/nodejs/"
	// srcPrefix := "stashes/check/"

	// err = CopyS3Folder(context.TODO(), s3Client, bucket, srcPrefix, dstPrefix)
	// if err != nil {
	// 	log.Fatalf("failed: %v", err)
	// }
	createStash("newtest","nodejs")
}
