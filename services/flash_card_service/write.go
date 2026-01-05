package flash_card_service

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/contract_resp"
	"github.com/umarkotak/ytkidd-api/repos/flash_card_repo"
)

func BulkInsert(ctx context.Context, params contract.BulkInsertFlashCard) (contract_resp.BulkInsertFlashCard, error) {
	ids, err := flash_card_repo.BulkInsert(ctx, nil, params.FlashCards)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.BulkInsertFlashCard{}, err
	}

	return contract_resp.BulkInsertFlashCard{
		IDs: ids,
	}, nil
}
