package utils_handler

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
)

const (
	CacheQuality = 75
)

func CompressHandler(w http.ResponseWriter, r *http.Request) {
	imgURL := r.URL.Query().Get("url")
	widthStr := r.URL.Query().Get("w")
	heightStr := r.URL.Query().Get("h")

	if imgURL == "" {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}

	// 2. Generate Unique Filename (Cache Key)
	// We hash the URL + dimensions so different sizes get different files.
	key := fmt.Sprintf("%s-%s-%s", imgURL, widthStr, heightStr)
	hash := fmt.Sprintf("%x", md5.Sum([]byte(key)))
	filename := fmt.Sprintf("%s.jpg", hash)
	cachePath := filepath.Join(config.Get().CacheDirPath, filename)

	// 3. CHECK DISK (Cache Hit)
	// If file exists, serve it directly.
	if _, err := os.Stat(cachePath); err == nil {
		// http.ServeFile handles ETag, Last-Modified, and Range requests automatically.
		w.Header().Set("X-Cache-Status", "HIT")
		http.ServeFile(w, r, cachePath)
		return
	}

	// 4. CACHE MISS - Fetch & Process
	fmt.Println("Cache MISS, processing:", imgURL)

	// Fetch from R2/Upstream
	resp, err := http.Get(imgURL)
	if err != nil {
		http.Error(w, "Failed to fetch upstream image", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Upstream image not found", http.StatusNotFound)
		return
	}

	// Decode
	srcImage, _, err := image.Decode(resp.Body)
	if err != nil {
		http.Error(w, "Failed to decode image", http.StatusInternalServerError)
		return
	}

	// Resize
	width, _ := strconv.Atoi(widthStr)
	height, _ := strconv.Atoi(heightStr)
	resizedImage := imaging.Resize(srcImage, width, height, imaging.Lanczos)

	// 5. SAVE TO DISK (Atomic Write)
	// We write to a temp file first, then rename it. This prevents users
	// from reading a file that is currently being written (corrupt image).
	tempFile, err := os.CreateTemp(config.Get().CacheDirPath, "temp-*.jpg")
	if err != nil {
		logrus.Error(err)
		http.Error(w, "Cache write error", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name()) // Clean up temp file if we crash before rename

	// Encode to the temp file
	if err := jpeg.Encode(tempFile, resizedImage, &jpeg.Options{Quality: CacheQuality}); err != nil {
		http.Error(w, "Encoding error", http.StatusInternalServerError)
		return
	}
	tempFile.Close()

	// Rename temp file to final cache path (Atomic operation)
	if err := os.Rename(tempFile.Name(), cachePath); err != nil {
		// On Windows, Rename fails if file exists. We can ignore or force it.
		// For simplicity, we log.
		log.Println("Could not rename cache file:", err)
	}

	// 6. SERVE THE NEW FILE
	w.Header().Set("X-Cache-Status", "MISS")
	http.ServeFile(w, r, cachePath)
}
