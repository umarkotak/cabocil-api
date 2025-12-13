package contract_resp

import (
	"time"
)

type (
	YoutubeChannel struct {
		ID         int64     `json:"id"`
		ExternalID string    `json:"external_id"`
		ImageUrl   string    `json:"image_url"`
		Name       string    `json:"name"`
		Tags       []string  `json:"tags"`
		Active     bool      `json:"active"`
		UpdatedAt  time.Time `json:"updated_at"`
	}

	YoutubeChannelDetailed struct {
		ID          int64     `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		ExternalID  string    `json:"external_id"`
		Name        string    `json:"name"`
		Username    string    `json:"username"`
		ImageUrl    string    `json:"image_url"`
		Tags        []string  `json:"tags"`
		Active      bool      `json:"active"`
		ChannelLink string    `json:"channel_link"`
	}
)
