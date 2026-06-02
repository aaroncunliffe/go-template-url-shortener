package links

type CreateLinkRequest struct {
	ShortPath string `validate:"omitempty,min=4,shortpath" json:"short_path"`
	OriginURL string `validate:"required,url" json:"origin_url"`
}

type CreateLinkResponse struct {
	ShortPath string `json:"short_path"`
}
