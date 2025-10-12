package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DefaultPerPage is the default number of items to return per page.
const DefaultPerPage = 10

// MaxPerPage is the maximum number of items that can be requested per page.
const MaxPerPage = 100

// Pagination contains the metadata for a paginated API response.
type Pagination struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	TotalPages  int   `json:"total_pages"`
	TotalItems  int64 `json:"total_items"`
}

// Params holds the validated pagination parameters extracted from a request.
type Params struct {
	Page    int
	PerPage int
	Offset  int
}

// GetPaginationParams extracts and validates pagination parameters from the Gin context.
func GetPaginationParams(c *gin.Context, defaultPerPage, maxPerPage int) Params {
	pageStr := c.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	perPageStr := c.Query("per_page")
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 {
		perPage = defaultPerPage
	}

	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	offset := (page - 1) * perPage

	return Params{
		Page:    page,
		PerPage: perPage,
		Offset:  offset,
	}
}

// NewPaginationMeta creates the pagination metadata object for an API response.
func NewPaginationMeta(p Params, totalItems int64) *Pagination {
	if p.PerPage <= 0 {
		return &Pagination{
			CurrentPage: p.Page,
			PerPage:     p.PerPage,
			TotalPages:  0,
			TotalItems:  totalItems,
		}
	}
	totalPages := int(math.Ceil(float64(totalItems) / float64(p.PerPage)))

	return &Pagination{
		CurrentPage: p.Page,
		PerPage:     p.PerPage,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
	}
}
