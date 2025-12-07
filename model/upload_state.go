package model

import "time"

type (
	UploadState struct {
		StatusMap map[string]UploadBookStatus `json:"status_map"`
	}

	UploadBookStatus struct {
		Slug        string    `json:"slug"`
		CreatedAt   time.Time `json:"created_at"`
		CurrentPage int       `json:"current_page"`
		MaxPage     int       `json:"max_page"`
	}
)
