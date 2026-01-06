package youtube_handler

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/services/youtube_service"
	"github.com/umarkotak/ytkidd-api/utils"
	"github.com/umarkotak/ytkidd-api/utils/render"
)

func ScrapVideos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := contract.ScrapVideos{}
	err := utils.BindJson(r, &params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	maxPage := params.MaxPage
	if maxPage <= 0 {
		maxPage = model.YOUTUBE_MAX_PAGE
	}

	if params.ChannelID == "" && params.ChannelUrl == "" {
		render.Error(w, r, err, "missing channel id or channel url")
		return
	}

	if params.ChannelID == "" && params.ChannelUrl != "" {
		params.ChannelID, err = youtube_service.GetChannelIDByChannelUrl(params.ChannelUrl)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			render.Error(w, r, err, "")
			return
		}
	}

	for i := 1; i <= int(maxPage); i++ {
		logrus.Infof("start scrap iteration: %v", i)

		nextPageToken, someVideoExist, err := youtube_service.ScrapVideos(ctx, params)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			render.Error(w, r, err, "")
			return
		}

		if params.BreakOnExists && someVideoExist {
			logrus.WithContext(ctx).Infof("iter: %v", i)
			break
		}

		params.PageToken = nextPageToken
		if params.PageToken == "" {
			logrus.Infof("break scrap iteration: %v", i)
			break
		}

		logrus.Infof("finish scrap iteration: %v", i)
	}

	render.Response(w, r, 200, params)
}
