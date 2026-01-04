package flash_card_repo

import (
	"context"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/model"
)

func GetByID(ctx context.Context, id int64) (model.FlashCard, error) {
	obj := model.FlashCard{}

	err := stmtGetByID.GetContext(ctx, &obj, map[string]any{
		"id": id,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return obj, err
	}

	return obj, nil
}

func GetByTags(ctx context.Context, tags pq.StringArray, limit, offset int64) ([]model.FlashCard, error) {
	objs := []model.FlashCard{}

	err := stmtGetByTags.SelectContext(ctx, &objs, map[string]any{
		"tags":   tags,
		"limit":  limit,
		"offset": offset,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return objs, err
	}

	return objs, nil
}
