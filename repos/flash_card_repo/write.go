package flash_card_repo

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/model"
)

func BulkInsert(ctx context.Context, tx *sqlx.Tx, flashCards []model.FlashCard) ([]int64, error) {
	var err error
	ids := []int64{}

	namedStmt := stmtBulkInsert
	if tx != nil {
		namedStmt, err = tx.PrepareNamedContext(ctx, queryBulkInsert)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return ids, err
		}
	}

	for _, fc := range flashCards {
		fmt.Printf("%+v", fc)
		var id int64
		err = namedStmt.GetContext(ctx, &id, fc)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return ids, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}
