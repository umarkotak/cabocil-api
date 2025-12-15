package user_handler

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/services/user_activity_service"
	"github.com/umarkotak/ytkidd-api/utils"
	"github.com/umarkotak/ytkidd-api/utils/common_ctx"
	"github.com/umarkotak/ytkidd-api/utils/render"
)

func GetUserActivities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	commonCtx := common_ctx.GetFromCtx(ctx)

	appSession := commonCtx.AppSession
	if commonCtx.UserAuth.GUID != "" {
		appSession = commonCtx.UserAuth.GUID
	}

	params := contract.GetUserActivity{
		UserGuid:   commonCtx.UserAuth.GUID,
		AppSession: appSession,
		Pagination: model.Pagination{
			Limit: utils.StringMustInt64(r.URL.Query().Get("limit")),
			Page:  utils.StringMustInt64(r.URL.Query().Get("page")),
		},
	}
	params.Pagination.SetDefault()

	data, err := user_activity_service.GetUserActivities(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, http.StatusOK, data)
}

func PostUserActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	commonCtx := common_ctx.GetFromCtx(ctx)

	appSession := commonCtx.AppSession
	if commonCtx.UserAuth.GUID != "" {
		appSession = commonCtx.UserAuth.GUID
	}

	params := contract.RecordUserActivity{
		UserGuid:   commonCtx.UserAuth.GUID,
		AppSession: appSession,
	}

	err := utils.BindJson(r, &params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	err = user_activity_service.Record(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, http.StatusOK, nil)
}

func AdminGetUserActivities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := contract.GetUserActivity{
		Pagination: model.Pagination{
			Limit: utils.StringMustInt64(r.URL.Query().Get("limit")),
			Page:  utils.StringMustInt64(r.URL.Query().Get("page")),
		},
	}
	params.Pagination.SetDefault()

	data, err := user_activity_service.GetRecentForAdmin(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, http.StatusOK, data)
}
