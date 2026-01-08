package book_link_handler

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/services/book_link_service"
	"github.com/umarkotak/ytkidd-api/utils/common_ctx"
	"github.com/umarkotak/ytkidd-api/utils/render"
)

func GetBookLinks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	commonCtx := common_ctx.GetFromCtx(ctx)

	bookLinkGroups, err := book_link_service.GetAllGrouped(ctx, commonCtx.UserAuth.GUID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, bookLinkGroups)
}
