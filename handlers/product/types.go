package product

import "github.com/YasserCherfaoui/MarketProGo/models"

// This file contains shared request/response structures for the product handlers.

type ImageData struct {
	FileName  string `json:"file_name"`
	IsPrimary bool   `json:"is_primary"`
	AltText   string `json:"alt_text"`
}

type ImageUpdateData struct {
	ID        uint    `json:"id"`
	IsPrimary *bool   `json:"is_primary"`
	AltText   *string `json:"alt_text"`
}

type SpecificationRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

type SpecificationUpdateData struct {
	ID    uint    `json:"id"`
	Name  *string `json:"name"`
	Value *string `json:"value"`
	Unit  *string `json:"unit"`
}

type OptionData struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type OptionUpdateData struct {
	ID     uint      `json:"id"`
	Name   *string   `json:"name"`
	Values *[]string `json:"values"`
}

type OptionDeleteData struct {
	ID uint `json:"id"`
}

type PriceTierData struct {
	MinQuantity int     `json:"min_quantity"`
	Price       float64 `json:"price"`
}

type VariantData struct {
	Name         string            `json:"name"`
	SKU          string            `json:"sku"`
	Barcode      string            `json:"barcode"`
	BasePrice    float64           `json:"base_price"`
	B2BPrice     float64           `json:"b2b_price"`
	CostPrice    float64           `json:"cost_price"`
	Weight       float64           `json:"weight"`
	WeightUnit   string            `json:"weight_unit"`
	Dimensions   models.Dimensions `json:"dimensions"`
	IsActive     bool              `json:"is_active"`
	Images       []ImageData       `json:"images"`
	OptionValues []string          `json:"option_values"`
	MinQuantity  int               `json:"min_quantity"`
	PriceTiers   []PriceTierData   `json:"price_tiers"`
}

type VariantUpdateData struct {
	ID             uint               `json:"id"`
	Name           *string            `json:"name"`
	SKU            *string            `json:"sku"`
	Barcode        *string            `json:"barcode"`
	BasePrice      *float64           `json:"base_price"`
	B2BPrice       *float64           `json:"b2b_price"`
	CostPrice      *float64           `json:"cost_price"`
	Weight         *float64           `json:"weight"`
	WeightUnit     *string            `json:"weight_unit"`
	Dimensions     *models.Dimensions `json:"dimensions"`
	IsActive       *bool              `json:"is_active"`
	ImagesToAdd    []ImageData        `json:"images_to_add"`
	ImagesToUpdate []ImageUpdateData  `json:"images_to_update"`
	ImagesToDelete []uint             `json:"images_to_delete"`
	OptionValues   *[]string          `json:"option_values"`
	MinQuantity    *int               `json:"min_quantity"`
	PriceTiers     *[]PriceTierData   `json:"price_tiers"`
}
