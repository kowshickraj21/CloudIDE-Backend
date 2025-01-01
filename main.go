package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"main/auth"
	"main/aws"
	"main/k8s"
	"main/ws"

	"net/http"
)

var db = &sql.DB{}


func init(){
	LoadEnv();
	db = ConnectDB();
}

func main() {

	client := AWSInit()
	if client == nil {
		log.Fatalln("Initialization Error!")
	}

	http.HandleFunc("/auth/google/callback",corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		user,err := auth.HandleGoogleUser(db,code)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	}))

	http.HandleFunc("/auth/github/callback", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		user,err := auth.HandleGithubUser(db,code) 
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	}))

	http.HandleFunc("/create", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var stash k8s.Stash
		err := json.NewDecoder(r.Body).Decode(&stash)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		bucket := os.Getenv("AWS_BUCKET")
		err = aws.CopyS3Folder(context.TODO(),client,bucket,fmt.Sprintf("new/%s",stash.Image),fmt.Sprintf("stashes/%s",stash.Name))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		_,err = k8s.CreateStash(db,stash)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		w.WriteHeader(http.StatusOK)
	}))

	http.HandleFunc("/run", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var details k8s.Stash
		err := json.NewDecoder(r.Body).Decode(&details)
		if err != nil {
			fmt.Println(err)
		}
		name,err := k8s.StartDeployment(details)
		fmt.Println(name)
		if err != nil {
			fmt.Println(err)
		}
		ws.StartSocket(w,r,client)
	}))


	if err := http.ListenAndServe(":3050", nil); err != nil {
		fmt.Println("Server error:", err)
	}
}



func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	origin := os.Getenv("CORS_ORIGIN")
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Origin, user")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}
