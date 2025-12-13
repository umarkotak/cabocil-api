package file_bucket

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/config"
	"github.com/umarkotak/ytkidd-api/datastore"
	"github.com/umarkotak/ytkidd-api/model"
)

func CreateFolderIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create the directory with appropriate permissions.
		// 0755 is a common permission setting (read/write/execute for owner, read/execute for others).
		err := os.MkdirAll(path, os.ModePerm) // Use MkdirAll to create parent directories if needed
		if err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}
		fmt.Println("Folder created:", path)
	} else if err != nil {
		return fmt.Errorf("checking if directory exists: %w", err)
	} else {
		fmt.Println("Folder already exists:", path)
	}
	return nil
}

func GenFileUrl(guid string) string {
	return fmt.Sprintf("%s/ytkidd/api/file_bucket/%s", config.Get().AppHost, guid)
}

func GenRawFileUrl(bucketName, fileBucketPath string) string {
	return fmt.Sprintf("%s/%s/%s", config.Get().AppHost, bucketName, fileBucketPath)
}

func GenFinalUrl(ctx context.Context, storageType, coverFilePath string) string {
	var coverFileUrl string
	if storageType == model.STORAGE_R2 {
		coverFileUrl, _ = datastore.GetObjectUrl(ctx, coverFilePath)
	} else {
		coverFileUrl = GenRawFileUrl(config.Get().FileBucketPath, coverFilePath)
	}
	return coverFileUrl
}

func GenCacheUrl(url, widht, height string) string {
	return fmt.Sprintf("%s/ytkidd/api/utils/image/compress?url=%s&w=%s&h=%s", config.Get().AppHost, url, widht, height)
}

func DeleteFileIfExists(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, which is fine.
			return nil
		}
		// Some other error occurred during deletion.
		return err
	}
	return nil
}

var errDangerousPath = errors.New("refusing to delete dangerous path")

// DeleteFolder deletes a directory and everything inside it.
// It refuses to delete "", ".", root, or the current working dir.
func DeleteFolder(path string) error {
	if path == "" || path == "." || path == "/" || path == `\` {
		logrus.Error(errDangerousPath)
		return errDangerousPath
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		logrus.Error(err)
		return err
	}

	// Extra guard: don't allow deleting the current working directory.
	cwd, err := os.Getwd()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if abs == cwd || abs == string(filepath.Separator) {
		logrus.Error(errDangerousPath)
		return errDangerousPath
	}

	// Optional: ensure it exists and is a directory (gives nicer errors)
	info, err := os.Lstat(abs)
	if err != nil {
		logrus.Error(err)
		return err
	}
	if !info.IsDir() {
		err = errors.New("path is not a directory")
		logrus.Error(err)
		return err
	}

	return os.RemoveAll(abs)
}

// CopyFile copies a file from src to dst.
// If dst does not exist, it will be created.
func CopyFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Ensure the file is flushed to disk
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to flush destination file: %w", err)
	}

	return nil
}
