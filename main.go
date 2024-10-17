package main

import (
	"context"
	"log"
	"os"

	"main/aws"
)

func main() {

	s3Client := AWSInit()
	if s3Client == nil {
		log.Fatalln("Initialization Error!")
	}

	bucket := os.Getenv("AWS_BUCKET")
	dstPrefix := "new/nodejs/"
	srcPrefix := "stashes/check/"

	err := aws.CopyS3Folder(context.TODO(), s3Client, bucket, srcPrefix, dstPrefix)
	if err != nil {
		log.Fatalf("failed: %v", err)
	}
}
