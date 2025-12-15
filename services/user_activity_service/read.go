package user_activity_service

import (
	"context"
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/contract_resp"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/user_activity_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_repo"
	"github.com/umarkotak/ytkidd-api/utils/file_bucket"
)

func GetUserActivities(ctx context.Context, params contract.GetUserActivity) (contract_resp.GetUserActivity, error) {
	var err error

	if params.UserGuid != "" {
		user, err := user_repo.GetByGuid(ctx, params.UserGuid)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return contract_resp.GetUserActivity{}, err
		}
		params.UserID = user.ID
	}

	if params.UserID == 0 && params.AppSession == "" {
		return contract_resp.GetUserActivity{}, fmt.Errorf("missing user identifier")
	}

	fullUserActivities, err := user_activity_repo.GetFullByParams(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.GetUserActivity{}, err
	}

	simpleUserActivities := make([]contract_resp.UserActivitySimple, 0, len(fullUserActivities))
	for _, fullUserActivity := range fullUserActivities {
		activityType := model.ACTIVITY_VIDEO
		if fullUserActivity.BookID != 0 {
			activityType = model.ACTIVITY_BOOK
		}

		userActivitySimple := contract_resp.UserActivitySimple{
			ActivityType:         activityType,
			YoutubeVideoID:       fullUserActivity.YoutubeVideoID,
			BookID:               fullUserActivity.BookID,
			BookContentID:        fullUserActivity.BookContentID,
			UserActivityMetadata: fullUserActivity.UserActivityMetadata,
		}

		if activityType == model.ACTIVITY_VIDEO {
			userActivitySimple.Video = contract_resp.UserActivityVideo{
				Title:           fullUserActivity.YoutubeVideoTitle.String,
				ImageUrl:        fullUserActivity.YoutubeVideoImageUrl.String,
				RedirectPath:    fmt.Sprintf("/watch/%v", fullUserActivity.YoutubeVideoID),
				ChannelName:     fullUserActivity.YoutubeChannelName.String,
				ChannelImageUrl: fullUserActivity.YoutubeChannelImageUrl.String,
			}

		} else {
			bookPageType := "books"
			if fullUserActivity.BookType.String == model.BOOK_TYPE_WORKBOOK {
				bookPageType = "workbooks"
			}
			imageUrl := file_bucket.GenCacheUrl(
				file_bucket.GenFinalUrl(ctx, fullUserActivity.BookCoverStorage.String, fullUserActivity.BookCoverExactPath.String),
				model.PRESET_COVER, 0, 0,
			)
			userActivitySimple.Book = contract_resp.UserActivityBook{
				Title:        fullUserActivity.BookTitle.String,
				ImageUrl:     imageUrl,
				RedirectPath: fmt.Sprintf("/%s/%s/read?page=%v", bookPageType, fullUserActivity.BookSlug.String, fullUserActivity.BookLastReadPageNumber.Int64),
			}
		}

		simpleUserActivities = append(simpleUserActivities, userActivitySimple)
	}

	return contract_resp.GetUserActivity{
		Activities: simpleUserActivities,
	}, nil
}

func GetRecentForAdmin(ctx context.Context, params contract.GetUserActivity) (contract_resp.GetRecentUserActivityForAdmin, error) {
	var err error

	recentUserActivities, err := user_activity_repo.GetRecentForAdmin(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.GetRecentUserActivityForAdmin{}, err
	}

	userMap := make(map[string]contract_resp.OneUserActivity)
	for _, fullUserActivity := range recentUserActivities {
		activityType := model.ACTIVITY_VIDEO
		if fullUserActivity.BookID != 0 {
			activityType = model.ACTIVITY_BOOK
		}

		userActivitySimple := contract_resp.UserActivitySimple{
			ActivityType:         activityType,
			YoutubeVideoID:       fullUserActivity.YoutubeVideoID,
			BookID:               fullUserActivity.BookID,
			BookContentID:        fullUserActivity.BookContentID,
			UserActivityMetadata: fullUserActivity.UserActivityMetadata,
		}

		if activityType == model.ACTIVITY_VIDEO {
			userActivitySimple.Video = contract_resp.UserActivityVideo{
				Title:           fullUserActivity.YoutubeVideoTitle.String,
				ImageUrl:        fullUserActivity.YoutubeVideoImageUrl.String,
				RedirectPath:    fmt.Sprintf("/watch/%v", fullUserActivity.YoutubeVideoID),
				ChannelName:     fullUserActivity.YoutubeChannelName.String,
				ChannelImageUrl: fullUserActivity.YoutubeChannelImageUrl.String,
			}

		} else {
			bookPageType := "books"
			if fullUserActivity.BookType.String == model.BOOK_TYPE_WORKBOOK {
				bookPageType = "workbooks"
			}
			imageUrl := file_bucket.GenCacheUrl(
				file_bucket.GenFinalUrl(ctx, fullUserActivity.BookCoverStorage.String, fullUserActivity.BookCoverExactPath.String),
				model.PRESET_COVER, 0, 0,
			)
			userActivitySimple.Book = contract_resp.UserActivityBook{
				Title:        fullUserActivity.BookTitle.String,
				ImageUrl:     imageUrl,
				RedirectPath: fmt.Sprintf("/%s/%s/read?page=%v", bookPageType, fullUserActivity.BookSlug.String, fullUserActivity.BookLastReadPageNumber.Int64),
			}
		}

		userKey := fmt.Sprintf("%v-%v", fullUserActivity.UserID, fullUserActivity.AppSession)
		if _, ok := userMap[userKey]; !ok {
			userMap[userKey] = contract_resp.OneUserActivity{
				Email:          fullUserActivity.UserEmail.String,
				AppSession:     fullUserActivity.AppSession,
				LastActivityAt: fullUserActivity.CreatedAt,
				Activities:     []contract_resp.UserActivitySimple{userActivitySimple},
			}
		} else {
			tempUser := userMap[userKey]
			tempUser.Activities = append(tempUser.Activities, userActivitySimple)
			userMap[userKey] = tempUser
		}
	}

	users := make([]contract_resp.OneUserActivity, 0, len(userMap))
	for _, user := range userMap {
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].LastActivityAt.After(users[j].LastActivityAt)
	})

	return contract_resp.GetRecentUserActivityForAdmin{
		Users: users,
	}, nil
}
