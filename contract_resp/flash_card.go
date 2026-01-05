package contract_resp

import "github.com/umarkotak/ytkidd-api/model"

type (
	GetFlashCard struct {
		FlashCard model.FlashCard `json:"flash_card"`
	}

	GetFlashCards struct {
		FlashCards []model.FlashCard `json:"flash_cards"`
	}

	BulkInsertFlashCard struct {
		IDs []int64 `json:"ids"`
	}
)
