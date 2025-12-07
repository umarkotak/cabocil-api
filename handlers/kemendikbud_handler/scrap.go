package kemendikbud_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/book_repo"
	"github.com/umarkotak/ytkidd-api/services/book_service"
)

type (
	// SourceResponse represents the structure of the input JSON.
	SourceResponse struct {
		Status  string         `json:"status"`
		Results []SourceResult `json:"results"`
	}

	// SourceResult represents each object in the "results" array of the input.
	SourceResult struct {
		Title       string `json:"title"`
		Slug        string `json:"slug"`
		Image       string `json:"image"`
		Attachment  string `json:"attachment"`
		Level       string `json:"level"`
		Subject     string `json:"subject"`
		Description string `json:"description"`
	}

	// TargetResponse represents the structure of the desired output JSON.
	TargetResponse struct {
		Books []TargetBook `json:"books"`
	}

	// TargetBook represents the desired structure for each book object.
	TargetBook struct {
		PdfURL      string   `json:"pdf_url"`
		Slug        string   `json:"slug"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		ImgFormat   string   `json:"img_format"`
		BookType    string   `json:"book_type"`
		Storage     string   `json:"storage"`
		StorePDF    bool     `json:"store_pdf"`
		Tags        []string `json:"tags"`
	}
)

var (
	TargetUrls = []string{
		"https://api.buku.cloudapp.web.id/api/catalogue/getTextBooks?limit=2000&type_pdf&order_by=updated_at",
		"https://api.buku.cloudapp.web.id/api/catalogue/getPenggerakTextBooks?limit=100&type_pdf&order_by=updated_at",
		"https://api.buku.cloudapp.web.id/api/catalogue/getNonTextBooks?limit=2000&type_pdf&&&tag=Buku%20Model",
	}

	TargetIndex = 2
)

func Scrap() {
	url := TargetUrls[TargetIndex]
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Unmarshal the input JSON into our source struct.
	var sourceData SourceResponse
	if err := json.Unmarshal(body, &sourceData); err != nil {
		log.Fatalf("Error unmarshaling source JSON: %v", err)
	}

	// Create the slice of target books.
	targetBooks := make([]TargetBook, 0, len(sourceData.Results))

	// Iterate over the results and transform each one.
	for _, item := range sourceData.Results {
		// Derive image format from the image URL extension.
		// imgExtension := strings.TrimPrefix(filepath.Ext(item.Image), ".")

		// Create the tags slice.
		tags := []string{
			fmt.Sprintf("level:%s", item.Level),
			fmt.Sprintf("subject:%s", item.Subject),
		}

		// Append the transformed book to our slice.
		targetBooks = append(targetBooks, TargetBook{
			PdfURL:      item.Attachment,
			Slug:        item.Slug,
			Title:       item.Title,
			Description: item.Description,
			ImgFormat:   model.IMAGE_JPEG,
			BookType:    model.BOOK_TYPE_DEFAULT, // Hardcoded value
			Storage:     "local",                 // Hardcoded value
			StorePDF:    false,                   // Hardcoded value
			Tags:        tags,
		})
	}

	ctx := context.Background()

	for idx, oneParams := range targetBooks {
		book, _ := book_repo.GetBySlug(ctx, oneParams.Slug)
		if book.ID > 0 {
			logrus.Infof("[%v/%v] %v already exists, skipping...", idx+1, len(targetBooks), oneParams.Title)
			continue
		}

		if oneParams.PdfURL == "" {
			logrus.Infof("[%v/%v] %v invalid pdf url, skipping...", idx+1, len(targetBooks), oneParams.Title)
			continue
		}

		httpClient := http.Client{
			Timeout: 30 * time.Minute,
		}
		resp, err := httpClient.Get(oneParams.PdfURL)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logrus.WithContext(ctx).Error(err)
			return
		}

		pdfBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return
		}

		if len(pdfBytes) == 0 {
			logrus.Infof("[%v/%v] %v invalid pdf file, skipping...", idx+1, len(targetBooks), oneParams.Title)
			continue
		}

		insertFromPdfParams := contract.InsertFromPdf{
			Slug:           oneParams.Slug,
			Title:          oneParams.Title,
			Description:    oneParams.Description,
			PdfBytes:       pdfBytes,
			ImgFormat:      oneParams.ImgFormat,
			BookType:       oneParams.BookType,
			OriginalPdfUrl: oneParams.PdfURL,
			Storage:        oneParams.Storage,
			StorePdf:       oneParams.StorePDF,
			Tags:           oneParams.Tags,
		}
		err = book_service.InsertFromPdf(ctx, insertFromPdfParams, &model.UploadState{})
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return
		}

		logrus.Infof("[%v/%v] %v added to library", idx+1, len(targetBooks), oneParams.Title)
	}
}
