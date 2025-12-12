package book_repo

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/utils"
)

func GetByID(ctx context.Context, id int64) (model.Book, error) {
	obj := model.Book{}
	err := stmtGetByID.GetContext(ctx, &obj, map[string]any{
		"id": id,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return obj, err
	}
	return obj, nil
}

func GetBySlug(ctx context.Context, slug string) (model.Book, error) {
	obj := model.Book{}
	err := stmtGetBySlug.GetContext(ctx, &obj, map[string]any{
		"slug": slug,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return obj, err
	}
	return obj, nil
}

func GetByIDOrSlug(ctx context.Context, identifier string) (model.Book, error) {
	var err error

	obj := model.Book{}

	id := utils.StringMustInt64(identifier)
	if id != 0 {
		err = stmtGetByID.GetContext(ctx, &obj, map[string]any{
			"id": id,
		})
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return obj, err
		}
		return obj, nil
	}

	err = stmtGetBySlug.GetContext(ctx, &obj, map[string]any{
		"slug": identifier,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return obj, err
	}
	return obj, nil
}

func GetByParams(ctx context.Context, params contract.GetBooks) ([]model.Book, error) {
	if params.Title != "" {
		params.Title = fmt.Sprintf("%%%s%%", params.Title)
	}
	if params.Tags == nil {
		params.Tags = []string{}
	}
	if params.Types == nil {
		params.Types = []string{}
	}
	if params.Access == nil {
		params.Access = []string{}
	}
	if params.ExcludeAccess == nil {
		params.ExcludeAccess = []string{}
	}
	if params.Active == "" {
		params.Active = "all"
	}

	objs := []model.Book{}
	err := stmtGetByParams.SelectContext(ctx, &objs, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return objs, err
	}
	return objs, nil
}

func GetTags(ctx context.Context) ([]string, error) {
	obj := pq.StringArray{}
	err := stmtGetTags.GetContext(ctx, &obj, map[string]any{})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return obj, err
	}
	return obj, nil
}
