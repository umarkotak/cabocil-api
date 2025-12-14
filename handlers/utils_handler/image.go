package utils_handler

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/datastore"
	"github.com/umarkotak/ytkidd-api/model"
)

func CompressHandler(w http.ResponseWriter, r *http.Request) {
	imgURL, widthStr, heightStr := datastore.GetCacheParam(r)

	if imgURL == "" {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}

	cachePath := datastore.GetCachePath(imgURL, widthStr, heightStr)

	// CHECK DISK (Cache Hit)
	if datastore.CheckCache(cachePath) {
		w.Header().Set(model.ImgCacheHeader, model.ImgCacheDuration7Days)
		w.Header().Set("X-Cache-Status", "HIT")
		http.ServeFile(w, r, cachePath)
		return
	}

	// CACHE MISS - Fetch & Process
	fmt.Println("Cache MISS, processing:", imgURL)
	err := datastore.SaveCache(imgURL, widthStr, heightStr, cachePath)
	if err != nil {
		logrus.Error(err)
		http.Error(w, "Failed to process image", http.StatusInternalServerError)
		return
	}

	// SERVE THE NEW FILE
	w.Header().Set(model.ImgCacheHeader, model.ImgCacheDuration7Days)
	w.Header().Set("X-Cache-Status", "MISS")
	http.ServeFile(w, r, cachePath)
}

func DeleteCacheHandler(w http.ResponseWriter, r *http.Request) {
	imgURL, widthStr, heightStr := datastore.GetCacheParam(r)

	if imgURL == "" {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}

	err := datastore.DeleteCache(imgURL, "", widthStr, heightStr)
	if err != nil {
		http.Error(w, "Failed to delete cache", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Cache deleted successfully"))
}
