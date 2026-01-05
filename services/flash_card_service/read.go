package flash_card_service

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/contract_resp"
	"github.com/umarkotak/ytkidd-api/repos/flash_card_repo"
)

func GetByID(ctx context.Context, params contract.GetFlashCardByID) (contract_resp.GetFlashCard, error) {
	flashCard, err := flash_card_repo.GetByID(ctx, params.ID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.GetFlashCard{}, err
	}

	return contract_resp.GetFlashCard{
		FlashCard: flashCard,
	}, nil
}

func GetByTags(ctx context.Context, params contract.GetFlashCardByTags) (contract_resp.GetFlashCards, error) {
	flashCards, err := flash_card_repo.GetByTags(ctx, params.Tags, params.Pagination.Limit, params.Pagination.Offset)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.GetFlashCards{}, err
	}

	return contract_resp.GetFlashCards{
		FlashCards: flashCards,
	}, nil
}
