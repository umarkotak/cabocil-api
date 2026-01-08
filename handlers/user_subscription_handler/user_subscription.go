package user_subscription_handler

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/services/user_subscription_service"
	"github.com/umarkotak/ytkidd-api/utils"
	"github.com/umarkotak/ytkidd-api/utils/render"
)

func DirectInject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := contract.UserSubscriptionDirectInject{}

	err := utils.BindJson(r, &params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	err = user_subscription_service.DirectInject(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, map[string]any{
		"message": "subscription injected successfully",
	})
}
