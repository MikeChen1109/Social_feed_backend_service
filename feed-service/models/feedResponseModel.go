package models

type PaginatedFeedsResponse struct {
	Data []Feed `json:"data"`
	Meta Meta   `json:"meta"`
}

type Meta struct {
	NextCursor string `json:"nextCursor,omitempty"`
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"hasMore"`
}
