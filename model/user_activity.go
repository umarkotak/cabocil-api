package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

const (
	ACTIVITY_VIDEO = "video"
	ACTIVITY_BOOK  = "book"
)

type (
	UserActivity struct {
		ID        int64        `db:"id"`
		CreatedAt time.Time    `db:"created_at"`
		UpdatedAt time.Time    `db:"updated_at"`
		DeletedAt sql.NullTime `db:"deleted_at"`

		UserID               int64                `db:"user_id"`
		AppSession           string               `db:"app_session"`
		YoutubeVideoID       int64                `db:"youtube_video_id"`
		BookID               int64                `db:"book_id"`
		BookContentID        int64                `db:"book_content_id"`
		UserActivityMetadata UserActivityMetadata `db:"metadata"`
	}

	UserActivityFull struct {
		ID        int64        `db:"id"`
		CreatedAt time.Time    `db:"created_at"`
		UpdatedAt time.Time    `db:"updated_at"`
		DeletedAt sql.NullTime `db:"deleted_at"`

		UserID               int64                `db:"user_id"`
		UserEmail            sql.NullString       `db:"user_email"`
		AppSession           string               `db:"app_session"`
		YoutubeVideoID       int64                `db:"youtube_video_id"`
		BookID               int64                `db:"book_id"`
		BookContentID        int64                `db:"book_content_id"`
		UserActivityMetadata UserActivityMetadata `db:"metadata"`

		YoutubeVideoTitle      sql.NullString `db:"youtube_video_title"`
		YoutubeVideoImageUrl   sql.NullString `db:"youtube_video_image_url"`
		YoutubeChannelName     sql.NullString `db:"youtube_channel_name"`
		YoutubeChannelImageUrl sql.NullString `db:"youtube_channel_image_url"`
		BookTitle              sql.NullString `db:"book_title"`
		BookCoverFileGuid      sql.NullString `db:"book_cover_file_guid"`
		BookCoverStorage       sql.NullString `db:"book_cover_storage"`
		BookCoverExactPath     sql.NullString `db:"book_cover_exact_path"`
		BookType               sql.NullString `db:"book_type"`
		BookSlug               sql.NullString `db:"book_slug"`
		BookLastReadPageNumber sql.NullInt64  `db:"book_last_read_page_number"`
	}

	UserActivityMetadata struct {
		LastReadBookContentID int64 `json:"last_read_book_content_id"`
		CurrentProgress       int64 `json:"current_progress"`
		MinProgress           int64 `json:"min_progress"`
		MaxProgress           int64 `json:"max_progress"`
	}
)

func (m UserActivityMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *UserActivityMetadata) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &m)
}
