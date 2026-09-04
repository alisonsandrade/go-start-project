// Package pagination
package pagination

import (
	"math"
	"net/http"
	"strconv"
)

type Params struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func ExtractParams(r *http.Request, defaultLimit, maxLimit int) Params {
	query := r.URL.Query()

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return Params{
		Page:  page,
		Limit: limit,
	}
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.Limit
}

type Meta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
}

type PageResult[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}

func NewPageResult[T any](items []T, total int64, params Params) PageResult[T] {
	totalPages := 0
	if params.Limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(params.Limit)))
	}

	return PageResult[T]{
		Data: items,
		Meta: Meta{
			CurrentPage: params.Page,
			PerPage:     params.Limit,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
	}
}
