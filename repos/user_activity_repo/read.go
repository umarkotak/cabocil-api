package user_activity_repo

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/model"
)

func GetByParams(ctx context.Context, params contract.GetUserActivity) ([]model.UserActivity, error) {
	objs := []model.UserActivity{}

	err := stmtGetByParams.SelectContext(ctx, &objs, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return objs, err
	}

	return objs, nil
}

func GetFullByParams(ctx context.Context, params contract.GetUserActivity) ([]model.UserActivityFull, error) {
	objs := []model.UserActivityFull{}

	err := stmtGetFullByParams.SelectContext(ctx, &objs, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return objs, err
	}

	return objs, nil
}

func GetByUserActivity(ctx context.Context, params contract.GetUserActivity) (model.UserActivity, error) {
	obj := model.UserActivity{}

	err := stmtGetByUserActivity.GetContext(ctx, &obj, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return obj, err
	}

	return obj, nil
}

func GetFullByUserActivity(ctx context.Context, params contract.GetUserActivity) (model.UserActivityFull, error) {
	obj := model.UserActivityFull{}

	err := stmtGetFullByUserActivity.GetContext(ctx, &obj, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return obj, err
	}

	return obj, nil
}

func GetRecentForAdmin(ctx context.Context, params contract.GetUserActivity) ([]model.UserActivityFull, error) {
	objs := []model.UserActivityFull{}

	err := stmtGetRecentForAdmin.SelectContext(ctx, &objs, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return objs, err
	}

	return objs, nil
}
