package flash_card_handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/services/flash_card_service"
	"github.com/umarkotak/ytkidd-api/utils"
	"github.com/umarkotak/ytkidd-api/utils/render"
)

func GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := utils.StringMustInt64(chi.URLParam(r, "id"))

	params := contract.GetFlashCardByID{
		ID: id,
	}

	data, err := flash_card_service.GetByID(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, http.StatusOK, data)
}

func GetByTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tagsParam := r.URL.Query().Get("tags")
	tags := pq.StringArray{}
	if tagsParam != "" {
		tags = pq.StringArray(strings.Split(tagsParam, ","))
	}

	params := contract.GetFlashCardByTags{
		Tags: tags,
		Pagination: model.Pagination{
			Limit: utils.StringMustInt64(r.URL.Query().Get("limit")),
			Page:  utils.StringMustInt64(r.URL.Query().Get("page")),
		},
	}
	params.Pagination.SetDefault()

	data, err := flash_card_service.GetByTags(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, http.StatusOK, data)
}

func BulkInsert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := contract.BulkInsertFlashCard{}

	err := utils.BindJson(r, &params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	data, err := flash_card_service.BulkInsert(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, http.StatusCreated, data)
}
