package contract

import (
	"mime/multipart"

	"github.com/lib/pq"
	"github.com/umarkotak/ytkidd-api/model"
)

type (
	InsertFromPdf struct {
		Slug           string
		Title          string
		Description    string
		PdfBytes       []byte
		ImgFormat      string // Enum: jpeg, png
		BookType       string
		OriginalPdfUrl string
		Storage        string // Enum: local, r2
		StorePdf       bool
		Tags           []string
	}

	InsertFromPdfUrl struct {
		PdfUrl      string   `json:"pdf_url"`
		Slug        string   `json:"slug"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		ImgFormat   string   `json:"img_format"`
		BookType    string   `json:"book_type"`
		Storage     string   `json:"storage"`
		StorePdf    bool     `json:"store_pdf"`
		Tags        []string `json:"tags"`
	}

	GetBooks struct {
		UserGuid      string         `db:"-"`
		UserRole      string         `db:"-"`
		Slug          string         `db:"-"`
		Title         string         `db:"title"`
		Tags          pq.StringArray `db:"tags"`
		Types         pq.StringArray `db:"types"`
		Access        pq.StringArray `db:"access"`
		ExcludeAccess pq.StringArray `db:"exclude_access"`
		Sort          string         `db:"sort"`
		ExcludeIDs    pq.Int64Array  `db:"exclude_ids"`
		Active        string         `db:"active"` // Enum: true, false, all
		model.Pagination
	}

	DeleteBook struct {
		BookID int64 `db:"id"`
	}

	RemoveBookPage struct {
		BookID         int64   `json:"-"`
		BookContentIDs []int64 `json:"book_content_ids"`
	}

	UpdateBook struct {
		ID             int64          `json:"-" db:"id"`
		Slug           string         `json:"slug" db:"slug"`
		Title          string         `json:"title" db:"title"`
		Description    string         `json:"description" db:"description"`
		Tags           pq.StringArray `json:"tags" db:"tags"`
		Type           string         `json:"type" db:"type"`
		Active         bool           `json:"active" db:"active"`
		OriginalPdfUrl string         `json:"original_pdf_url" db:"original_pdf_url"`
		AccessTags     pq.StringArray `json:"access_tags" db:"access_tags"`
	}

	UpdateBookCover struct {
		BookID    int64
		CoverFile multipart.File
	}
)
