package book_handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/book_repo"
	"github.com/umarkotak/ytkidd-api/services/book_service"
	"github.com/umarkotak/ytkidd-api/utils"
	"github.com/umarkotak/ytkidd-api/utils/render"
)

var (
	uploadState = model.UploadState{
		StatusMap: map[string]model.UploadBookStatus{},
	}
)

func GetBooksUploadStatus(w http.ResponseWriter, r *http.Request) {
	render.Response(w, r, 200, map[string]any{
		"status_map": uploadState.StatusMap,
	})
}

func InsertFromPdf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if _, exists := uploadState.StatusMap[r.FormValue("slug")]; exists {
		render.Error(w, r, fmt.Errorf("task is currently running for slug: %s", r.FormValue("slug")), "")
		return
	}

	err := r.ParseMultipartForm(model.PDF_MAX_FILE_SIZE_MB)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	pdfFile, _, err := r.FormFile("pdf_file")
	if err != nil && err != http.ErrMissingFile {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	pdfBytes := []byte{}
	if pdfFile != nil {
		defer pdfFile.Close()
		pdfBytes, err = io.ReadAll(pdfFile)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			render.Error(w, r, err, "")
			return
		}
	}

	if r.FormValue("pdf_url") != "" {
		httpClient := http.Client{}
		resp, err := httpClient.Get(r.FormValue("pdf_url"))
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			render.Error(w, r, err, "")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logrus.WithContext(ctx).Error(err)
			render.Error(w, r, err, "")
			return
		}

		pdfBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			render.Error(w, r, err, "")
			return
		}
	}

	if len(pdfBytes) == 0 {
		err = fmt.Errorf("pdf bytes empty")
		render.Error(w, r, err, "")
		return
	}
	params := contract.InsertFromPdf{
		Slug:        r.FormValue("slug"),
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		PdfBytes:    pdfBytes,
		ImgFormat:   r.FormValue("img_format"),
		BookType:    r.FormValue("book_type"),
		Storage:     r.FormValue("storage"),
		StorePdf:    r.FormValue("store_pdf") == "true",
		Tags:        utils.SplitString(r.FormValue("tags"), ","),
	}

	go func() {
		err = book_service.InsertFromPdf(context.Background(), params, &uploadState)
		if err != nil {
			logrus.WithContext(context.Background()).Error(err)
			return
		}
	}()

	render.Response(w, r, 200, nil)
}

func InsertFromPdfUrls(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := struct {
		Books []contract.InsertFromPdfUrl `json:"books"`
	}{}
	err := utils.BindJson(r, &params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	bgCtx := context.Background()
	go func() {
		for _, oneParams := range params.Books {
			book, _ := book_repo.GetBySlug(bgCtx, oneParams.Slug)
			if book.ID > 0 {
				continue
			}

			if oneParams.PdfUrl == "" {
				continue
			}

			uploadState.StatusMap[oneParams.Slug] = model.UploadBookStatus{
				Slug:      oneParams.Slug,
				CreatedAt: time.Now(),
			}
			defer func() {
				delete(uploadState.StatusMap, oneParams.Slug)
			}()

			httpClient := http.Client{}
			resp, err := httpClient.Get(oneParams.PdfUrl)
			if err != nil {
				logrus.WithContext(bgCtx).Error(err)
				// render.Error(w, r, err, "")
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				logrus.WithContext(bgCtx).Error(err)
				// render.Error(w, r, err, "")
				return
			}

			pdfBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				logrus.WithContext(bgCtx).Error(err)
				// render.Error(w, r, err, "")
				return
			}

			if len(pdfBytes) == 0 {
				continue
			}

			insertFromPdfParams := contract.InsertFromPdf{
				Slug:           oneParams.Slug,
				Title:          oneParams.Title,
				Description:    oneParams.Description,
				PdfBytes:       pdfBytes,
				ImgFormat:      oneParams.ImgFormat,
				BookType:       oneParams.BookType,
				OriginalPdfUrl: oneParams.PdfUrl,
				Storage:        oneParams.Storage,
				StorePdf:       oneParams.StorePdf,
				Tags:           oneParams.Tags,
			}
			err = book_service.InsertFromPdf(bgCtx, insertFromPdfParams, &uploadState)
			if err != nil {
				logrus.WithContext(bgCtx).Error(err)
				// render.Error(w, r, err, "")
				return
			}
		}
	}()

	render.Response(w, r, 200, nil)
}
