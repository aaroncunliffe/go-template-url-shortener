package links

type CreateLinkRequest struct {
	OriginURL string `validate:"required,url" json:"origin_url"`
}

type CreateLinkResponse struct {
	ShortPath string `json:"short_path"`
}
