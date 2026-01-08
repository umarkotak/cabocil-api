package user_subscription_service

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/user_subscription_repo"
)

func DirectInject(ctx context.Context, params contract.UserSubscriptionDirectInject) error {
	now := time.Now()
	endedAt := now.AddDate(0, 0, params.Days)

	subscription := model.UserSubscription{
		UserID:      params.UserID,
		OrderID:     -1,
		ProductCode: "direct_inject",
		StartedAt:   now,
		EndedAt:     endedAt,
	}

	_, err := user_subscription_repo.Insert(ctx, nil, subscription)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	return nil
}
