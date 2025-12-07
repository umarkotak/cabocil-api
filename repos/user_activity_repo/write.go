package user_activity_repo

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/model"
)

func Insert(ctx context.Context, tx *sqlx.Tx, obj model.UserActivity) (int64, error) {
	var err error

	namedStmt := stmtInsert
	if tx != nil {
		namedStmt, err = tx.PrepareNamedContext(ctx, queryInsert)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return 0, err
		}
	}

	err = namedStmt.GetContext(ctx, &obj.ID, obj)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return 0, err
	}

	return obj.ID, nil
}

func Upsert(ctx context.Context, tx *sqlx.Tx, obj model.UserActivity) (int64, error) {
	var err error

	namedStmt := stmtUpsert
	if tx != nil {
		namedStmt, err = tx.PrepareNamedContext(ctx, queryUpsert)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return 0, err
		}
	}

	err = namedStmt.GetContext(ctx, &obj.ID, obj)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return 0, err
	}

	return obj.ID, nil
}

func DeleteByYoutubeVideoIDs(ctx context.Context, tx *sqlx.Tx, youtubeVideoIDs pq.Int64Array) error {
	var err error

	namedStmt := stmtDeleteByYoutubeVideoIDs
	if tx != nil {
		namedStmt, err = tx.PrepareNamedContext(ctx, queryDeleteByYoutubeVideoIDs)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
	}

	_, err = namedStmt.ExecContext(ctx, map[string]any{
		"youtube_video_ids": youtubeVideoIDs,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	return nil
}

func DeleteByBookIDs(ctx context.Context, tx *sqlx.Tx, bookIDs pq.Int64Array) error {
	var err error

	namedStmt := stmtDeleteByBookIDs
	if tx != nil {
		namedStmt, err = tx.PrepareNamedContext(ctx, queryDeleteByBookIDs)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
	}

	_, err = namedStmt.ExecContext(ctx, map[string]any{
		"book_ids": bookIDs,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	return nil
}
