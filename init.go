package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func AWSInit() *s3.Client{
	staticCreds := aws.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY"),
			os.Getenv("AWS_ACCESS_SECRET"), ""))
			
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(os.Getenv("AWS_ACCESS_REGION")),
		config.WithCredentialsProvider(staticCreds),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
		return nil
	}

	s3Client := s3.NewFromConfig(cfg)
	fmt.Println("[AWS] Connected to AWS Server")
	return s3Client
}


func ConnectDB() (*sql.DB) {
	var err error

	connStr := os.Getenv("DB_URL")

	if IsStringEmpty(connStr){
		log.Fatalln("[DATABASE] Env Variables not found...")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Println(err)
		log.Fatalln("[DATABASE] Connection problem")
		return nil;
	}
	
	err = db.Ping();

	if err != nil {
		log.Println(err)
		log.Fatalln("[DATABASE] Could not ping the db.")
	}

	fmt.Println("[DATABASE] Connected to Database")
	return db;
}

func IsStringEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

func LoadEnv(){
	err := godotenv.Load()
	if err != nil{
		log.Fatal("No .env File");
	}
}