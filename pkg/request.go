package pkg

type PaginationRequest struct {
	Page  *int    `query:"page"`
	Limit *int    `query:"limit"`
	Sort  *string `query:"sort"`
}
