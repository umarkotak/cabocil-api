package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type (
	UserStroke struct {
		ID            int64        `db:"id"`
		CreatedAt     time.Time    `db:"created_at"`
		UpdatedAt     time.Time    `db:"updated_at"`
		DeletedAt     sql.NullTime `db:"deleted_at"`
		UserID        int64        `db:"user_id"`
		AppSession    string       `db:"app_session"`
		BookID        int64        `db:"book_id"`
		BookContentID int64        `db:"book_content_id"`
		ImageUrl      string       `db:"image_url"`
		Strokes       Strokes      `db:"strokes"`
	}

	Strokes []struct {
		Tool         string        `json:"tool"`
		Color        string        `json:"color"`
		RelativeSize float64       `json:"relativeSize"`
		Opacity      float64       `json:"opacity"`
		Points       []StrokePoint `json:"points"`
	}

	StrokePoint struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
)

func (m Strokes) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Strokes) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &m)
}
