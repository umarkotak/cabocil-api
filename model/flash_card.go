package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

type (
	FlashCard struct {
		ID        int64        `db:"id" json:"id"`
		CreatedAt time.Time    `db:"created_at" json:"created_at"`
		UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
		DeletedAt sql.NullTime `db:"deleted_at" json:"deleted_at"`

		Slug       string            `json:"slug" db:"slug"`
		NameID     string            `json:"name_id" db:"name_id"`
		NameEN     string            `json:"name_en" db:"name_en"`
		PictureURL string            `json:"picture_url" db:"picture_url"`
		Tags       pq.StringArray    `json:"tags" db:"tags"`
		Metadata   FlashCardMetadata `json:"metadata" db:"metadata"`
	}

	FlashCardMetadata struct {
		// Add metadata fields as needed
	}
)

func (m FlashCardMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *FlashCardMetadata) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &m)
}
