package youtube_service

import (
	"context"
	"database/sql"
	"html"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/config"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/youtube_channel_repo"
	"github.com/umarkotak/ytkidd-api/repos/youtube_video_repo"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func ScrapVideos(ctx context.Context, params contract.ScrapVideos) (string, bool, error) {
	nextPageToken := ""
	someVideoExist := false

	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(config.Get().YoutubeApiKey))
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return nextPageToken, someVideoExist, err
	}

	call := youtubeService.Search.List([]string{"id", "snippet"})
	call = call.ChannelId(params.ChannelID). //
							Q(params.Query).             //
							Type("video").               //
							PageToken(params.PageToken). //
							MaxResults(50).              // Get up to 50 results.
							Order("date")                // Newest videos first

	response, err := call.Do()
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return nextPageToken, someVideoExist, err
	}

	youtubeChannel, err := youtube_channel_repo.GetByExternalID(ctx, params.ChannelID)
	if err != nil && err != sql.ErrNoRows {
		logrus.WithContext(ctx).Error(err)
		return nextPageToken, someVideoExist, err
	}

	youtubeChannelDetail, err := GetYouTubeChannelDetails(config.Get().YoutubeApiKey, params.ChannelID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return nextPageToken, someVideoExist, err
	}

	if youtubeChannel.ID == 0 {
		youtubeChannel = model.YoutubeChannel{
			ExternalID:  params.ChannelID,
			Name:        youtubeChannelDetail.Name,
			Username:    youtubeChannelDetail.Name,
			ImageUrl:    youtubeChannelDetail.ThumbnailURL,
			Tags:        []string{},
			ChannelLink: youtubeChannelDetail.URL,
		}
		youtubeChannel.ID, err = youtube_channel_repo.Insert(ctx, nil, youtubeChannel)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return nextPageToken, someVideoExist, err
		}

	} else {
		youtubeChannel.Name = youtubeChannelDetail.Name
		youtubeChannel.Username = youtubeChannelDetail.Name
		youtubeChannel.ImageUrl = youtubeChannelDetail.ThumbnailURL
		err = youtube_channel_repo.Update(ctx, nil, youtubeChannel)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return nextPageToken, someVideoExist, err
		}
	}

	for _, item := range response.Items {
		if item.Id.Kind != "youtube#video" {
			continue
		}

		// js, _ := item.MarshalJSON()
		// logrus.Infof("VIDEO: %+v", string(js))

		youtubeVideo, err := youtube_video_repo.GetByExternalID(ctx, item.Id.VideoId)
		if err != nil && err != sql.ErrNoRows {
			logrus.WithContext(ctx).Error(err)
			return nextPageToken, someVideoExist, err
		}

		if youtubeVideo.ID != 0 {
			someVideoExist = true
			continue
		}

		publishedAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			logrus.WithContext(ctx).WithFields(logrus.Fields{
				"published_at": item.Snippet.PublishedAt,
			}).Error(err)
			publishedAt = time.Now()
		}
		youtubeVideo = model.YoutubeVideo{
			YoutubeChannelID: youtubeChannel.ID,
			ExternalId:       item.Id.VideoId,
			Title:            html.UnescapeString(item.Snippet.Title),
			ImageUrl:         item.Snippet.Thumbnails.Medium.Url,
			Tags:             []string{},
			Active:           true,
			PublishedAt:      publishedAt,
		}
		youtubeVideo.ID, err = youtube_video_repo.Insert(ctx, nil, youtubeVideo)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return nextPageToken, someVideoExist, err
		}
	}

	// respString, _ := response.MarshalJSON()
	// logrus.Infof("youtube response: %+v", string(respString))
	nextPageToken = response.NextPageToken

	return nextPageToken, someVideoExist, nil
}
