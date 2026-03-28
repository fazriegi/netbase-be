package pkg

type PaginationRequest struct {
	Page  *uint   `query:"page"`
	Limit *uint   `query:"limit"`
	Sort  *string `query:"sort"`
}
