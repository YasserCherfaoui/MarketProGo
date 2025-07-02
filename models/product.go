package models

import (
	"time"

	"gorm.io/gorm"
)

// Product represents the base product information.
type Product struct {
	gorm.Model
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	IsFeatured  bool   `gorm:"default:false" json:"is_featured"`
	BrandID     *uint  `json:"brand_id"`

	// Relationships
	Brand          *Brand                 `json:"brand" gorm:"foreignKey:BrandID"`
	Categories     []*Category            `gorm:"many2many:product_categories;" json:"categories"`
	Tags           []*Tag                 `gorm:"many2many:product_tags;" json:"tags"`
	Images         []ProductImage         `gorm:"foreignKey:ProductID" json:"images"`
	Options        []ProductOption        `gorm:"foreignKey:ProductID" json:"options"`
	Variants       []ProductVariant       `gorm:"foreignKey:ProductID" json:"variants"`
	Specifications []ProductSpecification `json:"specifications"`
}

// ProductVariant represents a specific version of a product, like size or color.
type ProductVariant struct {
	gorm.Model
	ProductID  uint        `json:"product_id"`
	Product    Product     `json:"product"`
	Name       string      `gorm:"not null" json:"name"` // e.g., "1kg", "500g", "250g"
	SKU        string      `gorm:"uniqueIndex;not null" json:"sku"`
	Barcode    string      `json:"barcode"`
	BasePrice  float64     `gorm:"not null" json:"base_price"`    // price for clients
	B2BPrice   float64     `json:"b2b_price"`                     // price for b2b customers
	CostPrice  float64     `json:"cost_price"`                    // cost price for the product
	Weight     float64     `json:"weight"`                        // weight of the product
	WeightUnit string      `json:"weight_unit"`                   // unit of weight
	Dimensions *Dimensions `gorm:"embedded" json:"dimensions"`    // dimensions of the product
	IsActive   bool        `gorm:"default:true" json:"is_active"` // if the variant is active

	// Relationships
	Images         []ProductImage        `gorm:"foreignKey:ProductVariantID" json:"images"`
	OptionValues   []*ProductOptionValue `gorm:"many2many:variant_option_values;" json:"option_values"`
	InventoryItems []InventoryItem       `json:"inventory_items"`
}

// ProductOption defines a configurable property for a product, like "Flavor" or "Size".
type ProductOption struct {
	gorm.Model
	ProductID uint                 `json:"product_id"`
	Name      string               `gorm:"not null" json:"name"` // e.g., "Flavor", "Size"
	Values    []ProductOptionValue `gorm:"foreignKey:ProductOptionID" json:"values"`
}

// ProductOptionValue represents a specific choice for a ProductOption, like "Orange" or "1L".
type ProductOptionValue struct {
	gorm.Model
	ProductOptionID uint   `json:"product_option_id"`
	Value           string `gorm:"not null" json:"value"` // e.g., "Orange", "1L"
}

// Dimensions represents the physical size of a product or variant.
type Dimensions struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Unit   string  `json:"unit"` // e.g., "cm", "in"
}

// Tag represents a keyword or label that can be associated with a product.
type Tag struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;not null" json:"name"`
}

// ProductImage can be associated with a base product or a specific variant.
type ProductImage struct {
	gorm.Model
	ProductID        *uint  `json:"product_id"` // Nullable for variant-specific images
	ProductVariantID *uint  `json:"product_variant_id"`
	URL              string `gorm:"not null" json:"url"`
	IsPrimary        bool   `gorm:"default:false" json:"is_primary"`
	AltText          string `json:"alt_text"`
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
	Products     []*Product  `gorm:"many2many:product_categories;" json:"products,omitempty"`
}

// InventoryItem tracks stock for a specific product variant in a warehouse.
type InventoryItem struct {
	gorm.Model
	ProductVariantID uint           `json:"product_variant_id"`
	ProductVariant   ProductVariant `json:"-"`
	WarehouseID      uint           `json:"warehouse_id"`
	Warehouse        Warehouse      `json:"warehouse"`
	Quantity         int            `gorm:"not null" json:"quantity"`
	Reserved         int            `gorm:"default:0" json:"reserved"`
	BatchNumber      string         `json:"batch_number"`
	ExpiryDate       *time.Time     `json:"expiry_date"`
	Status           string         `gorm:"default:'active'" json:"status"` // active, expired, damaged
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

// StockMovement tracks all inventory movements for audit purposes
type StockMovement struct {
	gorm.Model
	InventoryItemID uint          `json:"inventory_item_id"`
	InventoryItem   InventoryItem `json:"inventory_item"`
	MovementType    string        `gorm:"not null" json:"movement_type"` // adjustment_in, adjustment_out, transfer_in, transfer_out, sold, returned
	Quantity        int           `gorm:"not null" json:"quantity"`
	Reason          string        `json:"reason"`
	Notes           string        `json:"notes"`
	Reference       string        `json:"reference"` // Order ID, Transfer ID, etc.
	UserID          *uint         `json:"user_id"`
	User            *User         `json:"user,omitempty"`
}
