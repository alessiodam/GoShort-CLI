package structs

type LoginResponse struct {
	Message string `json:"message"`
	Session string `json:"session"`
	Success bool   `json:"success"`
	User    User   `json:"user"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
	Message  string `json:"message"`
	Success  bool   `json:"success"`
}
