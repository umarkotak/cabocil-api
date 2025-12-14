package datastore

import (
	"crypto/md5"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/config"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/utils"
)

const (
	CACHE_QUALITY = 90
)

func GetCacheParam(r *http.Request) (string, string, string) {
	imgURL := r.URL.Query().Get("url")
	widthStr := r.URL.Query().Get("w")
	heightStr := r.URL.Query().Get("h")
	preset := r.URL.Query().Get("preset")

	if preset != "" {
		if p, ok := model.Presets[preset]; ok {
			widthStr = strconv.Itoa(p.Width)
			heightStr = strconv.Itoa(p.Height)
		}
	}
	return imgURL, widthStr, heightStr
}

func GetCachePath(imgURL, widthStr, heightStr string) string {
	key := fmt.Sprintf("%s-%s-%s", imgURL, widthStr, heightStr)
	hash := fmt.Sprintf("%x", md5.Sum([]byte(key)))
	filename := fmt.Sprintf("%s.jpg", hash)
	return filepath.Join(config.Get().CacheDirPath, filename)
}

func CheckCache(cachePath string) bool {
	if _, err := os.Stat(cachePath); err == nil {
		return true
	}
	return false
}

func SaveCache(imgURL, widthStr, heightStr, cachePath string) error {
	// Fetch from R2/Upstream
	resp, err := http.Get(imgURL)
	if err != nil {
		return fmt.Errorf("failed to fetch upstream image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upstream image not found")
	}

	// Decode
	srcImage, _, err := image.Decode(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	// Resize
	width := utils.StringMustInt(widthStr)
	height := utils.StringMustInt(heightStr)
	resizedImage := imaging.Resize(srcImage, width, height, imaging.Lanczos)

	// SAVE TO DISK (Atomic Write)
	tempFile, err := os.CreateTemp(config.Get().CacheDirPath, "temp-*.jpg")
	if err != nil {
		return fmt.Errorf("cache write error: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up temp file if we crash before rename

	// Encode to the temp file
	if err := jpeg.Encode(tempFile, resizedImage, &jpeg.Options{Quality: CACHE_QUALITY}); err != nil {
		return fmt.Errorf("encoding error: %v", err)
	}
	tempFile.Close()

	// Rename temp file to final cache path (Atomic operation)
	if err := os.Rename(tempFile.Name(), cachePath); err != nil {
		log.Println("Could not rename cache file:", err)
		return err
	}
	return nil
}

func DeleteCache(imgURL, preset, widthStr, heightStr string) error {
	if preset != "" {
		if p, ok := model.Presets[preset]; ok {
			widthStr = strconv.Itoa(p.Width)
			heightStr = strconv.Itoa(p.Height)
		}
	}

	cachePath := GetCachePath(imgURL, widthStr, heightStr)
	if err := os.Remove(cachePath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted or doesn't exist
		}
		logrus.Error(err)
		return err
	}
	return nil
}
