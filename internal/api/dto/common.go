package dto

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValidationErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

type PaginatedRequest struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func (p *PaginatedRequest) Defaults() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 || p.Limit > 100 {
		p.Limit = 20
	}
}

func (p *PaginatedRequest) Offset() int {
	return (p.Page - 1) * p.Limit
}
