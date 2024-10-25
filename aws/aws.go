package aws

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

func CreateObject(ctx context.Context, client *s3.Client, bucket, object string) error {
	emptyBody := io.NopCloser(bytes.NewReader([]byte{}))
	_, err := client.PutObject(ctx,&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   emptyBody,
	})
	return err
}

func ListDirectory(ctx context.Context, client *s3.Client, bucket, prefix string) ([]string, error) {
    var contents []string
    input := &s3.ListObjectsV2Input{
        Bucket: aws.String(bucket),
        Prefix: aws.String(prefix),
    }

    for {
        result, err := client.ListObjectsV2(ctx, input)
        if err != nil {
			fmt.Println(err)
            return nil, err
        }
        for _, item := range result.Contents {
            contents = append(contents, *item.Key)
        }
        if result.IsTruncated != nil && *result.IsTruncated  {
            input.ContinuationToken = result.NextContinuationToken
        } else {
            break
        }
    }

    return contents, nil
}


func RenameFile(ctx context.Context, client *s3.Client, bucket, oldKey, newKey string) error {
		_, err := client.CopyObject(ctx,&s3.CopyObjectInput{
			Bucket:     aws.String(bucket),
			CopySource: aws.String(bucket + "/" + oldKey),
			Key:        aws.String(newKey),
		})
		return err
		// return DeleteFile(bucket, oldKey)
}


func DeleteS3Folder(ctx context.Context, client *s3.Client, bucket, folder string) error {
	listInput := &s3.ListObjectsV2Input{
        Bucket: aws.String(bucket),
        Prefix: aws.String(folder), 
    }

	listOutput, err := client.ListObjectsV2(context.TODO(), listInput)
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

    _, err = client.DeleteObjects(context.TODO(), deleteInput)
    if err != nil {
        log.Fatalf("Failed to delete objects: %v", err)
    }

    log.Println("Successfully deleted the folder and its contents.")
	return nil;
}

