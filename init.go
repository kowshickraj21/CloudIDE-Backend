package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

func AWSInit() *s3.Client{
	err := 	godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
		return nil
    }

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
	return s3Client
}