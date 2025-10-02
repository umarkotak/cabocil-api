package comfy_ui_handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/config"
	"github.com/umarkotak/ytkidd-api/utils/render"
)

const (
	ImgBasePath = "cabocil-api.cabocil.com/comfy_ui_gallery"
)

func Gallery(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(config.Get().ComfyUIOutputDir)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, r, err, "")
		return
	}

	imageUrls := []string{}
	for _, file := range files {
		if !file.IsDir() {
			imageUrls = append(imageUrls, fmt.Sprintf("%s/%s", ImgBasePath, file.Name()))
		}
	}

	render.Response(w, r, 200, map[string]any{
		"image_urls": imageUrls,
	})
}
