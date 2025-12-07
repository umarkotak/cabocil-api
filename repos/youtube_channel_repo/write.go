package youtube_channel_repo

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/model"
)

func Insert(ctx context.Context, tx *sqlx.Tx, youtubeChannel model.YoutubeChannel) (int64, error) {
	var err error
	newID := int64(0)

	stmt := stmtInsert
	if tx != nil {
		stmt, err = tx.PrepareNamedContext(ctx, queryInsert)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return newID, err
		}
	}

	err = stmt.GetContext(ctx, &newID, youtubeChannel)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return newID, err
	}

	return newID, nil
}

func Update(ctx context.Context, tx *sqlx.Tx, youtubeChannel model.YoutubeChannel) error {
	var err error

	stmt := stmtUpdate
	if tx != nil {
		stmt, err = tx.PrepareNamedContext(ctx, queryUpdate)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
	}

	_, err = stmt.ExecContext(ctx, youtubeChannel)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	return nil
}

func SoftDelete(ctx context.Context, tx *sqlx.Tx, id int64) error {
	var err error

	stmt := stmtSoftDelete
	if tx != nil {
		stmt, err = tx.PrepareNamedContext(ctx, querySoftDelete)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
	}

	_, err = stmt.ExecContext(ctx, map[string]any{
		"id": id,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	return nil
}

func Delete(ctx context.Context, tx *sqlx.Tx, id int64) error {
	var err error

	stmt := stmtDelete
	if tx != nil {
		stmt, err = tx.PrepareNamedContext(ctx, queryDelete)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
	}

	_, err = stmt.ExecContext(ctx, map[string]any{
		"id": id,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	return nil
}
