package youtube_video_repo

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/model"
)

func GetByID(ctx context.Context, id int64) (model.YoutubeVideo, error) {
	obj := model.YoutubeVideo{}
	err := stmtGetByID.GetContext(ctx, &obj, map[string]any{
		"id": id,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return obj, err
	}
	return obj, nil
}

func GetByExternalID(ctx context.Context, externalID string) (model.YoutubeVideo, error) {
	obj := model.YoutubeVideo{}
	err := stmtGetByExternalID.GetContext(ctx, &obj, map[string]any{
		"external_id": externalID,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return obj, err
	}
	return obj, nil
}

func GetForSearch(ctx context.Context, params contract.YoutubeVideoSearch) ([]model.YoutubeVideo, error) {
	objs := []model.YoutubeVideo{}
	err := stmtGetForSearch.SelectContext(ctx, &objs, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return objs, err
	}
	return objs, nil
}

func GetByParams(ctx context.Context, params contract.GetYoutubeVideos) ([]model.YoutubeVideoDetailed, error) {
	objs := []model.YoutubeVideoDetailed{}

	params.Pagination.SetDefault()
	if params.Tags == nil {
		params.Tags = []string{}
	}
	if params.ChannelIDs == nil {
		params.ChannelIDs = []int64{}
	}
	if params.ExcludeIDs == nil {
		params.ExcludeIDs = []int64{}
	}
	if params.ExcludeChannelIDs == nil {
		params.ExcludeChannelIDs = []int64{}
	}

	err := stmtGetByParams.SelectContext(ctx, &objs, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return objs, err
	}
	return objs, nil
}

func GetIDsByChannelID(ctx context.Context, youtubeChannelID int64) ([]int64, error) {
	objs := []int64{}

	err := stmtGetIDsByChannelID.SelectContext(ctx, &objs, map[string]any{
		"youtube_channel_id": youtubeChannelID,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return objs, err
	}
	return objs, nil
}
