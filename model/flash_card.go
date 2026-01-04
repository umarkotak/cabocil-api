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
		ID        int64        `db:"id"`
		CreatedAt time.Time    `db:"created_at"`
		UpdatedAt time.Time    `db:"updated_at"`
		DeletedAt sql.NullTime `db:"deleted_at"`

		Slug       string            `db:"slug"`
		NameID     string            `db:"name_id"`
		NameEN     string            `db:"name_en"`
		PictureURL string            `db:"picture_url"`
		Tags       pq.StringArray    `db:"tags"`
		Metadata   FlashCardMetadata `db:"metadata"`
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
