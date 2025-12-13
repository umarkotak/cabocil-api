package youtube_channel_service

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/contract_resp"
	"github.com/umarkotak/ytkidd-api/datastore"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/user_activity_repo"
	"github.com/umarkotak/ytkidd-api/repos/youtube_channel_repo"
	"github.com/umarkotak/ytkidd-api/repos/youtube_video_repo"
)

func GetChannels(ctx context.Context, params contract.GetYoutubeChannels) ([]contract_resp.YoutubeChannel, error) {
	respYoutubeChannels := []contract_resp.YoutubeChannel{}

	youtubeChannels, err := youtube_channel_repo.GetForSearch(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return respYoutubeChannels, err
	}

	for _, youtubeChannel := range youtubeChannels {
		respYoutubeChannels = append(respYoutubeChannels, contract_resp.YoutubeChannel{
			ID:         youtubeChannel.ID,
			ImageUrl:   youtubeChannel.ImageUrl,
			Name:       youtubeChannel.Name,
			Tags:       youtubeChannel.Tags,
			ExternalID: youtubeChannel.ExternalID,
			Active:     youtubeChannel.Active,
			UpdatedAt:  youtubeChannel.UpdatedAt,
		})
	}

	return respYoutubeChannels, nil
}

func UpdateChannel(ctx context.Context, params contract.UpdateYoutubeChannel) error {
	youtubeChannel, err := youtube_channel_repo.GetByID(ctx, params.ID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	youtubeChannel.ExternalID = params.ExternalID
	youtubeChannel.Name = params.Name
	youtubeChannel.Username = params.Username
	youtubeChannel.ImageUrl = params.ImageUrl
	youtubeChannel.Active = params.Active
	youtubeChannel.ChannelLink = params.ChannelLink

	err = youtube_channel_repo.Update(ctx, nil, youtubeChannel)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	return nil
}

func UpdateChannelActive(ctx context.Context, params contract.UpdateYoutubeChannel) error {
	err := youtube_channel_repo.UpdateActive(ctx, nil, model.YoutubeChannel{
		ID:     params.ID,
		Active: params.Active,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	return nil
}

func DeleteChannel(ctx context.Context, youtubeChannelID int64) error {
	youtubeChannel, err := youtube_channel_repo.GetByID(ctx, youtubeChannelID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	youtubeVideoIDs, err := youtube_video_repo.GetIDsByChannelID(ctx, youtubeChannel.ID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	err = datastore.Transaction(ctx, datastore.Get().Db, func(tx *sqlx.Tx) error {
		err = user_activity_repo.DeleteByYoutubeVideoIDs(ctx, tx, youtubeVideoIDs)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}

		err = youtube_video_repo.DeleteByChannelID(ctx, tx, youtubeChannel.ID)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}

		err = youtube_channel_repo.Delete(ctx, tx, youtubeChannel.ID)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	return nil
}
