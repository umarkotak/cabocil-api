package model

import (
	"database/sql"
	"encoding/json"
	"time"
)

type (
	BookLink struct {
		ID        int64           `db:"id"`
		CreatedAt time.Time       `db:"created_at"`
		UpdatedAt time.Time       `db:"updated_at"`
		DeletedAt sql.NullTime    `db:"deleted_at"`
		GroupName string          `db:"group_name"`
		Name      string          `db:"name"`
		Url       string          `db:"url"`
		ImageUrl  string          `db:"image_url"`
		Premium   bool            `db:"premium"`
		Metadata  json.RawMessage `db:"metadata"`
	}
)
