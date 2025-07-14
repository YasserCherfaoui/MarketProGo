# Database Models

This document describes the main database models used in the backend, their fields, and relationships. All models use [GORM](https://gorm.io/) for ORM mapping.

---

## Product

```go
// Product represents the base product information.
type Product struct {
    gorm.Model
    Name        string
    Description string
    IsActive    bool
    IsFeatured  bool
    IsVAT       bool
    BrandID     *uint
    Brand          *Brand
    Categories     []*Category
    Tags           []*Tag
    Images         []ProductImage
    Options        []ProductOption
    Variants       []ProductVariant
    Specifications []ProductSpecification
}
```
- **Purpose:** Main product entity, with relationships to brand, categories, tags, images, options, variants, and specifications.

---

## ProductVariant

```go
type ProductVariant struct {
    gorm.Model
    ProductID  uint
    Product    Product
    Name       string
    SKU        string
    Barcode    string
    BasePrice  float64
    B2BPrice   float64
    CostPrice  float64
    Weight     float64
    WeightUnit string
    Dimensions *Dimensions
    IsActive   bool
    Images         []ProductImage
    OptionValues   []*ProductOptionValue
    InventoryItems []InventoryItem
    MinQuantity int `gorm:"default:1" json:"min_quantity"` // minimum quantity to buy
    PriceTiers  []ProductVariantPriceTier `gorm:"foreignKey:ProductVariantID" json:"price_tiers"`
}
```
- **Purpose:** Represents a specific version of a product (e.g., size, color), with inventory, option values, minimum quantity, and dynamic pricing tiers.

---

## ProductVariantPriceTier

```go
type ProductVariantPriceTier struct {
    gorm.Model
    ProductVariantID uint    `json:"product_variant_id"`
    MinQuantity      int     `gorm:"not null" json:"min_quantity"` // minimum quantity for this price tier
    Price            float64 `gorm:"not null" json:"price"`
}
```
- **Purpose:** Represents a price break for a variant based on quantity (dynamic pricing/quantity discounts).

---

## Dynamic Pricing & Quantity Discounts

- Each product variant can have multiple price tiers, each specifying a minimum quantity and a price.
- When adding to cart or placing an order, the system selects the correct price tier based on the requested quantity.
- The `min_quantity` field enforces the minimum allowed quantity for purchase.
- Example:

| Quantity Range | Price  |
|---------------|--------|
| 1–10          | £10.00 |
| 11–50         | £9.00  |
| 51+           | £8.50  |

---

## Brand

```go
type Brand struct {
    gorm.Model
    Name        string
    Image       string
    Slug        string
    IsDisplayed bool
    ParentID *uint
    Parent   *Brand
    Children []*Brand
}
```
- **Purpose:** Product brand, supports parent-child relationships for brand hierarchies.

---

## Category

```go
type Category struct {
    gorm.Model
    Name         string
    Slug         string
    Description  string
    Image        string
    ParentID     *uint
    Parent       *Category
    IsFeatureOne bool
    Children     []*Category
    Products     []*Product
}
```
- **Purpose:** Product category, supports nesting and many-to-many with products.

---

## Tag

```go
type Tag struct {
    gorm.Model
    Name string
}
```
- **Purpose:** Keyword or label for products.

---

## ProductImage

```go
type ProductImage struct {
    gorm.Model
    ProductID        *uint
    ProductVariantID *uint
    URL              string
    IsPrimary        bool
    AltText          string
}
```
- **Purpose:** Image for a product or variant.

---

## InventoryItem

```go
type InventoryItem struct {
    gorm.Model
    ProductVariantID uint
    ProductVariant   ProductVariant
    WarehouseID      uint
    Warehouse        Warehouse
    Quantity         int
    Reserved         int
    BatchNumber      string
    ExpiryDate       *time.Time
    Status           string
}
```
- **Purpose:** Tracks stock for a product variant in a warehouse, with batch and expiry info.

---

## Warehouse

```go
type Warehouse struct {
    gorm.Model
    Name           string
    Code           string
    AddressID      uint
    Address        Address
    IsActive       bool
    InventoryItems []InventoryItem
}
```
- **Purpose:** Physical warehouse for inventory.

---

## ProductSpecification

```go
type ProductSpecification struct {
    gorm.Model
    ProductID uint
    Product   Product
    Name      string
    Value     string
    Unit      string
}
```
- **Purpose:** Key-value specifications for a product.

---

## StockMovement

```go
type StockMovement struct {
    gorm.Model
    InventoryItemID uint
    InventoryItem   InventoryItem
    MovementType    string
    Quantity        int
    Reason          string
    Notes           string
    Reference       string
    UserID          *uint
    User            *User
}
```
- **Purpose:** Tracks all inventory movements for audit and reporting.

---

## Order

```go
type Order struct {
    gorm.Model
    OrderNumber    string
    UserID         uint
    User           User
    CompanyID      *uint
    Company        *Company
    Status         OrderStatus
    PaymentStatus  PaymentStatus
    TotalAmount    float64
    TaxAmount      float64
    ShippingAmount float64
    DiscountAmount float64
    FinalAmount    float64
    ShippingAddressID uint
    ShippingAddress   Address
    ShippingMethod    string
    TrackingNumber    string
    PaymentMethod    string
    PaymentReference string
    PaymentDate      *time.Time
    Items []OrderItem
    CustomerNotes string
    AdminNotes    string
    OrderDate     time.Time
    ShippedDate   *time.Time
    DeliveredDate *time.Time
}
```
- **Purpose:** Customer or admin order, with items, shipping, payment, and status info.

---

## OrderItem

```go
type OrderItem struct {
    gorm.Model
    OrderID uint
    Order   Order
    ProductVariantID uint
    ProductVariant   ProductVariant
    ProductID *uint
    Product   *Product
    Quantity       int
    UnitPrice      float64
    TaxAmount      float64
    DiscountAmount float64
    TotalAmount    float64
    InventoryItemID *uint
    InventoryItem   *InventoryItem
    Status string
}
```
- **Purpose:** Line item in an order, linked to a product variant and inventory.

---

## Invoice

```go
type Invoice struct {
    gorm.Model
    OrderID          uint
    Order            Order
    InvoiceNumber    string
    IssueDate        time.Time
    DueDate          time.Time
    Amount           float64
    TaxAmount        float64
    Status           string
    PaymentDate      *time.Time
    PaymentMethod    string
    PaymentReference string
    Notes            string
}
```
- **Purpose:** Invoice for an order, with payment and status info.

---

## User

```go
type User struct {
    gorm.Model
    Email     string
    Password  string
    FirstName string
    LastName  string
    Phone     string
    UserType  UserType
    IsActive  bool
    LastLogin time.Time
    CompanyID *uint
    Role      string
    Addresses []*Address
    Company   *Company
}
```
- **Purpose:** User account, with support for B2B, roles, and addresses.

---

## Company

```go
type Company struct {
    gorm.Model
    Name               string
    VATNumber          string
    RegistrationNumber string
    Phone              string
    Email              string
    Website            string
    IsVerified         bool
    CreditLimit        float64
    PaymentTerms       int
    AddressID uint
    Users []*User
}
```
- **Purpose:** B2B company, with users and address.

---

## Address

```go
type Address struct {
    gorm.Model
    StreetAddress1 string
    StreetAddress2 string
    City           string
    State          string
    PostalCode     string
    Country        string
    IsDefault      bool
    UserID *uint
    User   *User
}
```
- **Purpose:** Address for users and companies.

---

## Promotion

```go
type Promotion struct {
    gorm.Model
    Title       string
    Description string
    Image       string
    ButtonText  string
    ButtonLink  string
    StartDate   time.Time
    EndDate     time.Time
    IsActive    bool
    ProductID  *uint
    Product    *Product
    CategoryID *uint
    Category   *Category
    BrandID    *uint
    Brand      *Brand
}
```
- **Purpose:** Marketing promotion, optionally linked to product, category, or brand.

---

For more details on each model, see the corresponding Go source files in the `models/` directory. 