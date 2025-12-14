package model

import (
	"database/sql"
	"slices"
	"time"

	"github.com/lib/pq"
)

const (
	ACCESS_TAG_FREE    = "free"
	ACCESS_TAG_PREMIUM = "premium"

	IMAGE_JPEG = "jpeg"
	IMAGE_PNG  = "png"

	BOOK_TYPE_DEFAULT  = "default"
	BOOK_TYPE_WORKBOOK = "workbook"
)

type (
	Book struct {
		ID             int64          `db:"id"`
		CreatedAt      time.Time      `db:"created_at"`
		UpdatedAt      time.Time      `db:"updated_at"`
		DeletedAt      sql.NullTime   `db:"deleted_at"`
		Slug           string         `db:"slug"`
		Title          string         `db:"title"`
		Description    string         `db:"description"`
		CoverFileGuid  string         `db:"cover_file_guid"`
		Tags           pq.StringArray `db:"tags"`
		Type           string         `db:"type"`
		PdfFileGuid    string         `db:"pdf_file_guid"`
		Active         bool           `db:"active"`
		OriginalPdfUrl string         `db:"original_pdf_url"`
		AccessTags     pq.StringArray `db:"access_tags"`
		Storage        string         `db:"storage"`

		CoverFilePath string        `db:"cover_file_path"` // join from file bucket
		ContentCount  sql.NullInt64 `db:"content_count"`   // join from book contents
	}
)

func (m Book) IsFree() bool {
	return slices.Contains(m.AccessTags, ACCESS_TAG_FREE)
}
