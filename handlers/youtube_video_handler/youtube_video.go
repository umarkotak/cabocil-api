package youtube_video_handler

import (
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/youtube_video_repo"
	"github.com/umarkotak/ytkidd-api/services/youtube_video_service"
	"github.com/umarkotak/ytkidd-api/utils"
	"github.com/umarkotak/ytkidd-api/utils/common_ctx"
	"github.com/umarkotak/ytkidd-api/utils/render"
)

func GetYoutubeVideos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	commonCtx := common_ctx.GetFromCtx(ctx)
	active := "true"
	if slices.Contains(model.ADMIN_ROLES, commonCtx.UserAuth.UserRole) {
		active = "all"
	}

	params := contract.GetYoutubeVideos{
		UserGuid:          commonCtx.UserAuth.GUID,
		UserRole:          commonCtx.UserAuth.UserRole,
		Tags:              utils.StringMustSliceString(r.URL.Query().Get("tags"), ","),
		ChannelIDs:        utils.StringMustSliceInt64(r.URL.Query().Get("channel_ids"), ","),
		ExcludeIDs:        utils.StringMustSliceInt64(r.URL.Query().Get("exclude_ids"), ","),
		ExcludeChannelIDs: utils.StringMustSliceInt64(r.URL.Query().Get("exclude_channel_ids"), ","),
		Sort:              r.URL.Query().Get("sort"),
		Active:            active,
		Pagination: model.Pagination{
			Limit: utils.StringMustInt64(r.URL.Query().Get("limit")),
			Page:  utils.StringMustInt64(r.URL.Query().Get("page")),
		},
	}

	youtubeVideosHome, err := youtube_video_service.GetVideos(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, youtubeVideosHome)
}

func GetYoutubeVideoDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	youtubeVideoID := utils.StringMustInt64(chi.URLParam(r, "id"))

	youtubeVideoDetail, err := youtube_video_service.GetVideoDetail(ctx, youtubeVideoID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, youtubeVideoDetail)
}

func DeleteVideo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	youtubeVideoID := utils.StringMustInt64(chi.URLParam(r, "id"))

	err := youtube_video_repo.Delete(ctx, nil, youtubeVideoID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, nil)
}
