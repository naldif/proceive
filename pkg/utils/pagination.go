package utils

import (
	"math"
	"strconv"
	"strings"

	"proceive/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

type Pagination struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	SortBy string `json:"sort_by,omitempty"`
	Order  string `json:"order,omitempty"`
	Search string `json:"search,omitempty"`
}

func ParsePagination(c *gin.Context) *Pagination {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = DefaultPage
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 {
		limit = DefaultLimit
	} else if limit > MaxLimit {
		limit = MaxLimit
	}

	sortBy := strings.TrimSpace(c.DefaultQuery("sort_by", "created_at"))
	if sortBy == "" {
		sortBy = "created_at"
	}

	order := strings.ToUpper(strings.TrimSpace(c.DefaultQuery("order", "DESC")))
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	search := strings.TrimSpace(c.Query("search"))
	if search == "" {
		search = strings.TrimSpace(c.Query("q"))
	}

	return &Pagination{
		Page:   page,
		Limit:  limit,
		SortBy: sortBy,
		Order:  order,
		Search: search,
	}
}

func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

func (p *Pagination) OrderClause() string {
	return p.SortBy + " " + p.Order
}

func (p *Pagination) Scope() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(p.Offset()).Limit(p.Limit).Order(p.OrderClause())
	}
}

func (p *Pagination) Meta(total int64) response.PaginationMeta {
	totalPages := 0
	if p.Limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(p.Limit)))
	}

	return response.PaginationMeta{
		Page:      p.Page,
		Limit:     p.Limit,
		Total:     total,
		TotalPage: totalPages,
		HasNext:   p.Page < totalPages,
		HasPrev:   p.Page > 1,
	}
}

// Aliases for compatibility
func GetPagination(c *gin.Context) *Pagination { return ParsePagination(c) }
func (p *Pagination) GormPaginate() func(db *gorm.DB) *gorm.DB { return p.Scope() }
func (p *Pagination) GetOffset() int                           { return p.Offset() }
func (p *Pagination) GetLimit() int                            { return p.Limit }
func (p *Pagination) GetOrderClause() string                   { return p.OrderClause() }
func (p *Pagination) BuildMeta(total int64) response.PaginationMeta {
	return p.Meta(total)
}
