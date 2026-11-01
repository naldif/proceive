package item

import (
	"time"

	"github.com/google/uuid"
)

type CreateItemRequest struct {
	Code          string  `json:"code" binding:"required,max=50"`
	Name          string  `json:"name" binding:"required,max=150"`
	Category      string  `json:"category" binding:"required,max=100"`
	UOM           string  `json:"uom" binding:"required,max=20"`
	PriceEstimate float64 `json:"price_estimate" binding:"gte=0"`
}

type UpdateItemRequest struct {
	Code          string   `json:"code" binding:"omitempty,max=50"`
	Name          string   `json:"name" binding:"omitempty,max=150"`
	Category      string   `json:"category" binding:"omitempty,max=100"`
	UOM           string   `json:"uom" binding:"omitempty,max=20"`
	PriceEstimate *float64 `json:"price_estimate" binding:"omitempty,gte=0"`
}

type ItemResponse struct {
	ID            uuid.UUID `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Category      string    `json:"category"`
	UOM           string    `json:"uom"`
	PriceEstimate float64   `json:"price_estimate"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func ToItemResponse(i *Item) ItemResponse {
	return ItemResponse{
		ID:            i.ID,
		Code:          i.Code,
		Name:          i.Name,
		Category:      i.Category,
		UOM:           i.UOM,
		PriceEstimate: i.PriceEstimate,
		CreatedAt:     i.CreatedAt,
		UpdatedAt:     i.UpdatedAt,
	}
}

func ToItemResponseList(items []Item) []ItemResponse {
	res := make([]ItemResponse, len(items))
	for idx, itm := range items {
		res[idx] = ToItemResponse(&itm)
	}
	return res
}
