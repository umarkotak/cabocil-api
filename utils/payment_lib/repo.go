package payment_lib

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func GetByID(ctx context.Context, id int64) (Payment, error) {
	obj := Payment{}
	err := stmtGetByID.GetContext(ctx, &obj, map[string]any{
		"id": id,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return obj, err
	}
	return obj, nil
}

func GetByNumber(ctx context.Context, number string) (Payment, error) {
	obj := Payment{}
	err := stmtGetByNumber.GetContext(ctx, &obj, map[string]any{
		"number": number,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return obj, err
	}
	return obj, nil
}

func GetByOrderNumber(ctx context.Context, orderNumber string) (Payment, error) {
	obj := Payment{}
	err := stmtGetByOrderNumber.GetContext(ctx, &obj, map[string]any{
		"order_number": orderNumber,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return obj, err
	}
	return obj, nil
}

func Insert(ctx context.Context, tx *sqlx.Tx, payment Payment) (int64, string, error) {
	var err error

	randomUUID, _ := uuid.NewRandom()
	payment.Number = randomUUID.String()

	namedStmt := stmtInsert
	if tx != nil {
		namedStmt, err = tx.PrepareNamedContext(ctx, queryInsert)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return 0, "", err
		}
	}

	err = namedStmt.GetContext(ctx, &payment.ID, payment)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return 0, "", err
	}

	return payment.ID, payment.Number, nil
}

func Update(ctx context.Context, tx *sqlx.Tx, payment Payment) error {
	var err error

	namedStmt := stmtUpdate
	if tx != nil {
		namedStmt, err = tx.PrepareNamedContext(ctx, queryUpdate)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
	}

	_, err = namedStmt.ExecContext(ctx, payment)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	return nil
}
