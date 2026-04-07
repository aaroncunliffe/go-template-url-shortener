package links

type CreateLinkRequest struct {
	ShortPath string `validate:"required,min=4" json:"short_path"`
	OriginURL string `validate:"required,url" json:"origin_url"`
}

type CreateLinkResponse struct {
	ShortPath string `json:"short_path"`
}
