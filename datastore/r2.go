package datastore

import (
	"bytes"
	"context"
	"encoding/json"
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

	// --- NEW LOGIC: Purge Cache ---
	// After successful upload, trigger the cache purge for this file.
	// You might want to log the error rather than failing the whole request
	// if the upload itself succeeded.
	if purgeErr := purgeCloudflareCache(ctx, objectKey); purgeErr != nil {
		// Log error only, so we don't return an error for a successful upload
		fmt.Printf("Warning: Failed to purge cache for %s: %v\n", objectKey, purgeErr)
	}
	// ------------------------------

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

type CloudflarePurgeReq struct {
	Files []string `json:"files"`
}

func purgeCloudflareCache(ctx context.Context, objectKey string) error {
	// 1. Construct the full public URL of the file
	// Assuming config.Get().PublicDomain is something like "https://cdn.example.com"
	fullURL := fmt.Sprintf("%s/%s", config.Get().R2PublicDomain, objectKey)

	// 2. Prepare the payload
	payload := CloudflarePurgeReq{
		Files: []string{fullURL},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	// 3. Create the HTTP Request to Cloudflare API
	// API Endpoint: https://api.cloudflare.com/client/v4/zones/{zone_identifier}/purge_cache
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/purge_cache", config.Get().CloudflareZoneId)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// 4. Set Headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Get().R2TokenValue)

	// 5. Execute Request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// 6. Check Response Status
	if resp.StatusCode != http.StatusOK {
		// You can read the body here to see the specific Cloudflare error if needed
		return fmt.Errorf("cloudflare api returned status: %d", resp.StatusCode)
	}

	return nil
}
