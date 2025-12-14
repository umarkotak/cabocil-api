package contract_resp

type (
	GetBooks struct {
		TagGroup []TagGroup `json:"tag_group"`
		Books    []Book     `json:"books"`
	}

	TagGroup struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	Book struct {
		ID           int64    `json:"id"`
		Slug         string   `json:"slug"`
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		CoverFileUrl string   `json:"cover_file_url"`
		Tags         []string `json:"tags"`
		Type         string   `json:"type"`
		IsFree       bool     `json:"is_free"`
		Active       bool     `json:"active"`
		Storage      string   `json:"storage"`
		ContentCount int64    `json:"content_count"`
	}

	BookDetail struct {
		ID           int64         `json:"id"`
		Slug         string        `json:"slug"`
		Title        string        `json:"title"`
		Description  string        `json:"description"`
		CoverFileUrl string        `json:"cover_file_url"`
		ThumbnailUrl string        `json:"thumbnail_url"`
		Tags         []string      `json:"tags"`
		Type         string        `json:"type"`
		Contents     []BookContent `json:"contents"`
		Active       bool          `json:"active"`
		AccessTags   []string      `json:"access_tags"`
		PdfUrl       string        `json:"pdf_url"`
		CanAction    bool          `json:"can_action"`
		IsFree       bool          `json:"is_free"`
		IsSubscribed bool          `json:"is_subscribed"`
		MaxPage      int           `json:"max_page"`
	}

	BookContent struct {
		ID           int64  `json:"id"`
		BookID       int64  `json:"book_id"`
		PageNumber   int64  `json:"page_number"`
		ImageFileUrl string `json:"image_file_url"`
		ThumbnailUrl string `json:"thumbnail_url"`
		Description  string `json:"description"`
	}
)
