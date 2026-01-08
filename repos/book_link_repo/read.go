package book_link_repo

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/model"
)

func GetAll(ctx context.Context) ([]model.BookLink, error) {
	objs := []model.BookLink{}
	err := stmtGetAll.SelectContext(ctx, &objs, map[string]any{})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return objs, err
	}
	return objs, nil
}
