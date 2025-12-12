package book_handler

import (
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/services/book_service"
	"github.com/umarkotak/ytkidd-api/utils"
	"github.com/umarkotak/ytkidd-api/utils/common_ctx"
	"github.com/umarkotak/ytkidd-api/utils/render"
)

func GetBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	commonCtx := common_ctx.GetFromCtx(ctx)
	active := "true"
	if slices.Contains(model.ADMIN_ROLES, commonCtx.UserAuth.UserRole) {
		active = "all"
	}

	params := contract.GetBooks{
		Title:         r.URL.Query().Get("title"),
		Tags:          utils.StringMustSliceString(r.URL.Query().Get("tags"), ","),
		Types:         utils.StringMustSliceString(r.URL.Query().Get("types"), ","),
		Access:        utils.StringMustSliceString(r.URL.Query().Get("access"), ","),
		ExcludeAccess: utils.StringMustSliceString(r.URL.Query().Get("exclude_access"), ","),
		Sort:          r.URL.Query().Get("sort"),
		ExcludeIDs:    utils.StringMustSliceInt64(r.URL.Query().Get("exclude_ids"), ","),
		Active:        active,
		Pagination: model.Pagination{
			Limit: utils.StringMustInt64(r.URL.Query().Get("limit")),
			Page:  utils.StringMustInt64(r.URL.Query().Get("page")),
		},
	}
	if params.Sort == "" {
		params.Sort = "title_asc"
	}
	params.Pagination.SetDefault()

	getBooks, err := book_service.GetBooks(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, getBooks)
}

func GetBookDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	commonCtx := common_ctx.GetFromCtx(ctx)

	params := contract.GetBooks{
		UserGuid: commonCtx.UserAuth.GUID,
		UserRole: commonCtx.UserAuth.UserRole,
		Slug:     chi.URLParam(r, "slug"),
	}

	bookDetail, err := book_service.GetBookDetail(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, bookDetail)
}

func DeleteBook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := contract.DeleteBook{
		BookID: utils.StringMustInt64(chi.URLParam(r, "id")),
	}

	err := book_service.DeleteBook(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, nil)
}

func RemoveBookPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := contract.RemoveBookPage{
		BookID: utils.StringMustInt64(chi.URLParam(r, "id")),
	}

	err := utils.BindJson(r, &params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	err = book_service.RemoveBookPage(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, nil)
}

func UpdateBook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := contract.UpdateBook{
		ID: utils.StringMustInt64(chi.URLParam(r, "id")),
	}

	err := utils.BindJson(r, &params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	err = book_service.UpdateBook(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, nil)
}

func UpdateBookCover(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := r.ParseMultipartForm(model.PDF_MAX_FILE_SIZE_MB)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	coverFile, _, err := r.FormFile("cover_file")
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	params := contract.UpdateBookCover{
		BookID:    utils.StringMustInt64(chi.URLParam(r, "id")),
		CoverFile: coverFile,
	}

	err = book_service.UpdateBookCover(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, 200, nil)
}
