package contract_resp

type (
	BookLinkGroup struct {
		GroupName string     `json:"group_name"`
		BookLinks []BookLink `json:"book_links"`
	}

	BookLink struct {
		Name     string `json:"name"`
		Url      string `json:"url"`
		ImageUrl string `json:"image_url"`
		Premium  bool   `json:"premium"`
	}
)
