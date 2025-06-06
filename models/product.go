package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name        string  `gorm:"not null" json:"name"`
	Description string  `json:"description"`
	SKU         string  `gorm:"uniqueIndex;not null" json:"sku"`
	Barcode     string  `json:"barcode"`
	QRCode      string  `json:"qr_code"`                    // URL to stored QR code image
	BasePrice   float64 `gorm:"not null" json:"base_price"` // Price for the customer
	B2BPrice    float64 `json:"b2b_price"`                  // Price for the business
	CostPrice   float64 `json:"cost_price"`                 // Price for the supplier
	Weight      float64 `json:"weight"`                     // Weight of the product
	WeightUnit  string  `json:"weight_unit"`                // Unit of weight
	IsActive    bool    `gorm:"default:true" json:"is_active"`
	IsFeatured  bool    `gorm:"default:false" json:"is_featured"`

	// Images
	Images []ProductImage `json:"images"`

	// Categories
	Categories []Category `gorm:"many2many:product_categories;" json:"categories"`

	// Inventory
	InventoryItems []InventoryItem `json:"inventory_items"`

	// Specifications
	Specifications []ProductSpecification `json:"specifications"`
}

type ProductImage struct {
	gorm.Model
	ProductID uint   `json:"product_id"`
	URL       string `gorm:"not null" json:"url"`
	IsPrimary bool   `gorm:"default:false" json:"is_primary"`
	AltText   string `json:"alt_text"`
}

type Category struct {
	gorm.Model
	Name         string      `gorm:"not null" json:"name"`
	Slug         string      `gorm:"uniqueIndex;not null" json:"slug"`
	Description  string      `json:"description"`
	Image        string      `json:"image"`
	ParentID     *uint       `json:"parent_id"`
	Parent       *Category   `json:"parent,omitempty"`
	IsFeatureOne bool        `gorm:"default:false" json:"is_feature_one"`
	Children     []*Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Products     []*Product  `gorm:"many2many:product_categories;" json:"products"`
}

type InventoryItem struct {
	gorm.Model
	ProductID   uint      `json:"product_id"`
	Product     Product   `json:"-"`
	WarehouseID uint      `json:"warehouse_id"`
	Warehouse   Warehouse `json:"warehouse"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	Reserved    int       `gorm:"default:0" json:"reserved"`
	BatchNumber string    `json:"batch_number"`
	ExpiryDate  time.Time `json:"expiry_date"`
	Status      string    `gorm:"default:'active'" json:"status"` // active, expired, damaged
}

type Warehouse struct {
	gorm.Model
	Name           string          `gorm:"not null" json:"name"`
	Code           string          `gorm:"uniqueIndex;not null" json:"code"`
	AddressID      uint            `json:"address_id"`
	Address        Address         `json:"address"`
	IsActive       bool            `gorm:"default:true" json:"is_active"`
	InventoryItems []InventoryItem `json:"inventory_items"`
}

type ProductSpecification struct {
	gorm.Model
	ProductID uint    `json:"product_id"`
	Product   Product `json:"-"`
	Name      string  `gorm:"not null" json:"name"`
	Value     string  `gorm:"not null" json:"value"`
	Unit      string  `json:"unit"`
}
