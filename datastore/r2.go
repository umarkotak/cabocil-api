package datastore

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/umarkotak/ytkidd-api/config"
)

func UploadFileToR2(ctx context.Context, filePath, objectKey string, deleteAfterUpload bool, cacheSecond uint) (err error) {
	if filePath == "" {
		return fmt.Errorf("filePath is required")
	}
	if objectKey == "" {
		return fmt.Errorf("objectKey is required")
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	// Defer closure to handle cleanup and error updating
	defer func() {
		// Always close the file first
		cerr := f.Close()

		// If the main logic succeeded but closing failed, return the close error
		if err == nil && cerr != nil {
			err = fmt.Errorf("close file: %w", cerr)
			return
		}

		// Only delete if requested AND no prior error (upload & close succeeded)
		if deleteAfterUpload && err == nil {
			if remErr := os.Remove(filePath); remErr != nil {
				err = fmt.Errorf("remove file: %w", remErr)
			}
		}
	}()

	// Peek first 512 bytes to detect content-type
	header := make([]byte, 512)
	n, _ := io.ReadFull(f, header)
	contentType := http.DetectContentType(header[:n])

	// Reset reader to the start
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek file: %w", err)
	}

	_, err = dataStore.R2Manager.Upload(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(config.Get().R2BucketName),
		Key:          aws.String(objectKey),
		Body:         f,
		ContentType:  aws.String(contentType),
		ACL:          types.ObjectCannedACLPublicRead,
		CacheControl: aws.String(fmt.Sprintf("max-age=%d", cacheSecond)),
	})
	if err != nil {
		return fmt.Errorf("upload: %w", err)
	}

	return nil
}

func GetPresignedUrl(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	req, err := dataStore.R2PresignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(config.Get().R2BucketName),
		Key:    aws.String(objectKey),
		// Example: set response headers if you want forced downloads
		// ResponseContentDisposition: aws.String("attachment"),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("presign: %w", err)
	}

	return req.URL, nil
}

func GetObjectUrl(ctx context.Context, objectKey string) (string, error) {
	return dataStore.R2PublicDomain + "/" + objectKey, nil
}

func DeleteByKeys(ctx context.Context, keys []string) error {
	// Build a batch of keys to delete
	objects := make([]types.ObjectIdentifier, 0, len(keys))
	for _, key := range keys {
		objects = append(objects, types.ObjectIdentifier{
			Key: &key,
		})
	}

	_, err := dataStore.R2Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(config.Get().R2BucketName),
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   true,
		},
	})
	if err != nil {
		return fmt.Errorf("delete objects: %w", err)
	}

	return nil
}

// DeleteObjectsByPrefix deletes all objects in the bucket that match the given prefix.
func DeleteObjectsByPrefix(ctx context.Context, prefix string) error {
	if prefix == "" {
		return fmt.Errorf("prefix is required")
	}

	// First, list objects with the prefix
	paginator := s3.NewListObjectsV2Paginator(dataStore.R2Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(config.Get().R2BucketName),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list objects: %w", err)
		}

		if len(page.Contents) == 0 {
			continue
		}

		// Build a batch of keys to delete
		var objects []types.ObjectIdentifier
		for _, obj := range page.Contents {
			objects = append(objects, types.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		_, err = dataStore.R2Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(config.Get().R2BucketName),
			Delete: &types.Delete{
				Objects: objects,
				Quiet:   true,
			},
		})
		if err != nil {
			return fmt.Errorf("delete objects: %w", err)
		}
	}

	return nil
}
