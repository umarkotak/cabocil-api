package youtube_channel_handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/contract_resp"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/youtube_channel_repo"
	"github.com/umarkotak/ytkidd-api/services/youtube_channel_service"
	"github.com/umarkotak/ytkidd-api/services/youtube_video_service"
	"github.com/umarkotak/ytkidd-api/utils"
	"github.com/umarkotak/ytkidd-api/utils/common_ctx"
	"github.com/umarkotak/ytkidd-api/utils/render"
)

func GetYoutubeChannels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := contract.GetYoutubeChannels{
		Name: r.URL.Query().Get("name"),
		Tags: utils.StringMustSliceString(r.URL.Query().Get("tags"), ","),
	}

	youtubeChannels, err := youtube_channel_service.GetChannels(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, youtubeChannels)
}

func GetYoutubeChannelDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	commonCtx := common_ctx.GetFromCtx(ctx)

	youtubeChannelID := utils.StringMustInt64(chi.URLParam(r, "id"))

	youtubeChannel, err := youtube_channel_repo.GetByID(ctx, youtubeChannelID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	youtubeVideos, err := youtube_video_service.GetVideos(ctx, contract.GetYoutubeVideos{
		UserGuid:   commonCtx.UserAuth.GUID,
		UserRole:   commonCtx.UserAuth.UserRole,
		ChannelIDs: []int64{youtubeChannelID},
		Sort:       r.URL.Query().Get("sort"),
		Pagination: model.Pagination{
			Limit: utils.StringMustInt64(r.URL.Query().Get("limit")),
			Page:  utils.StringMustInt64(r.URL.Query().Get("page")),
		},
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, map[string]any{
		"channel": contract_resp.YoutubeChannel{
			ID:       youtubeChannel.ID,
			ImageUrl: youtubeChannel.ImageUrl,
			Name:     youtubeChannel.Name,
			Tags:     youtubeChannel.Tags,
		},
		"videos": youtubeVideos.Videos,
	})
}

func GetYoutubeChannelDetailed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	youtubeChannelID := utils.StringMustInt64(chi.URLParam(r, "id"))

	youtubeChannel, err := youtube_channel_repo.GetByID(ctx, youtubeChannelID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200,
		contract_resp.YoutubeChannelDetailed{
			ID:          youtubeChannel.ID,
			CreatedAt:   youtubeChannel.CreatedAt,
			UpdatedAt:   youtubeChannel.UpdatedAt,
			ExternalID:  youtubeChannel.ExternalID,
			Name:        youtubeChannel.Name,
			Username:    youtubeChannel.Username,
			ImageUrl:    youtubeChannel.ImageUrl,
			Tags:        youtubeChannel.Tags,
			Active:      youtubeChannel.Active,
			ChannelLink: youtubeChannel.ChannelLink,
		},
	)
}

func UpdateYoutubeChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := contract.UpdateYoutubeChannel{
		ID: utils.StringMustInt64(chi.URLParam(r, "id")),
	}
	err := utils.BindJson(r, &params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	err = youtube_channel_service.UpdateChannel(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, map[string]any{})
}

func DeleteYoutubeChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := youtube_channel_service.DeleteChannel(ctx, utils.StringMustInt64(chi.URLParam(r, "id")))
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, map[string]any{})
}
