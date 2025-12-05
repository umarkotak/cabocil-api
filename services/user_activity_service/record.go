package user_activity_service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/book_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_activity_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_repo"
	"github.com/umarkotak/ytkidd-api/repos/youtube_video_repo"
)

func Record(ctx context.Context, params contract.RecordUserActivity) error {
	var err error

	if params.UserGuid != "" {
		user, err := user_repo.GetByGuid(ctx, params.UserGuid)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
		params.UserID = user.ID
	}

	if params.UserID == 0 && params.AppSession == "" {
		return fmt.Errorf("missing user identifier")
	}

	if params.BookID == 0 && params.BookContentID == 0 && params.YoutubeVideoID == 0 {
		return fmt.Errorf("missing content identifier")
	}

	if params.BookID != 0 {
		_, err = book_repo.GetByID(ctx, params.BookID)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
	}

	if params.YoutubeVideoID != 0 {
		_, err = youtube_video_repo.GetByID(ctx, params.YoutubeVideoID)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
	}

	_, err = user_activity_repo.Upsert(ctx, nil, model.UserActivity{
		UserID:               params.UserID,
		AppSession:           params.AppSession,
		YoutubeVideoID:       params.YoutubeVideoID,
		BookID:               params.BookID,
		BookContentID:        params.BookContentID,
		UserActivityMetadata: params.UserActivityMetadata,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	return nil
}
