package models

type PaginatedFeedsResponse struct {
	Data []Feed `json:"data"`
	Meta Meta   `json:"meta"`
}

type Meta struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	HasMore bool `json:"hasMore"`
}
