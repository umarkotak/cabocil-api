package contract

import (
	"github.com/lib/pq"
	"github.com/umarkotak/ytkidd-api/model"
)

type (
	GetFlashCardByID struct {
		ID int64 `db:"id"`
	}

	GetFlashCardByTags struct {
		Tags pq.StringArray `db:"tags"`
		model.Pagination
	}

	BulkInsertFlashCard struct {
		FlashCards []model.FlashCard `json:"flash_cards"`
	}
)
