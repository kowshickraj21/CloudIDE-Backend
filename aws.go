package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func CopyS3Folder(ctx context.Context, s3Client *s3.Client, bucket, srcPrefix, dstPrefix string) error {
	listObjectsInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(srcPrefix),
	}

	paginator := s3.NewListObjectsV2Paginator(s3Client, listObjectsInput)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects: %v", err)
		}

		for _, object := range page.Contents {
			srcKey := *object.Key
			dstKey := strings.Replace(srcKey, srcPrefix, dstPrefix, 1)

			copySource := fmt.Sprintf("%s/%s", bucket, srcKey)
			_, err := s3Client.CopyObject(ctx, &s3.CopyObjectInput{
				Bucket:     aws.String(bucket),
				CopySource: aws.String(copySource),
				Key:        aws.String(dstKey),
			})
			if err != nil {
				return fmt.Errorf("failed to copy object: %v", err)
			}

			fmt.Printf("Copied %s to %s\n", srcKey, dstKey)
		}
	}

	return nil
}

func DeleteS3Folder(ctx context.Context, s3Client *s3.Client, bucket, folder string) error {
	listInput := &s3.ListObjectsV2Input{
        Bucket: aws.String(bucket),
        Prefix: aws.String(folder), 
    }

	listOutput, err := s3Client.ListObjectsV2(context.TODO(), listInput)
    if err != nil {
        log.Fatalf("Failed to list objects in folder: %v", err)
    }

    var objectsToDelete []types.ObjectIdentifier
    for _, object := range listOutput.Contents {
        objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
            Key: object.Key,
        })
    }

    if len(objectsToDelete) == 0 {
        log.Println("No objects found to delete in folder.")
        return nil
    }

    deleteInput := &s3.DeleteObjectsInput{
        Bucket: aws.String(bucket),
        Delete: &types.Delete{
            Objects: objectsToDelete,
            Quiet:   aws.Bool(false),
        },
    }

    _, err = s3Client.DeleteObjects(context.TODO(), deleteInput)
    if err != nil {
        log.Fatalf("Failed to delete objects: %v", err)
    }

    log.Println("Successfully deleted the folder and its contents.")
	return nil;
}

