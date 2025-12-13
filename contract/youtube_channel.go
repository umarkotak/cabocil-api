package contract

import (
	"github.com/lib/pq"
)

type (
	GetYoutubeChannels struct {
		Name   string         `db:"name"`
		Tags   pq.StringArray `db:"tags"`
		Active string         `db:"active"`
	}

	UpdateYoutubeChannel struct {
		ID          int64  `json:"id"`
		ExternalID  string `json:"external_id"`
		Name        string `json:"name"`
		Username    string `json:"username"`
		ImageUrl    string `json:"image_url"`
		Active      bool   `json:"active"`
		ChannelLink string `json:"channel_link"`
	}
)
