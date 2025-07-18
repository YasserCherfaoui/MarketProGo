package models

import (
	"time"

	"gorm.io/gorm"
)

// ReviewStatus represents the moderation status of a product review
type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "PENDING"
	ReviewStatusApproved ReviewStatus = "APPROVED"
	ReviewStatusRejected ReviewStatus = "REJECTED"
	ReviewStatusFlagged  ReviewStatus = "FLAGGED"
)

// ProductReview represents a customer review for a specific product variant
type ProductReview struct {
	gorm.Model
	ProductVariantID uint           `json:"product_variant_id" gorm:"index"`
	ProductVariant   ProductVariant `json:"product_variant"`
	UserID           uint           `json:"user_id" gorm:"index"`
	User             User           `json:"user"`
	OrderItemID      *uint          `json:"order_item_id" gorm:"index"` // For purchase verification
	OrderItem        *OrderItem     `json:"order_item,omitempty"`

	Rating             int    `json:"rating" validate:"required,min=1,max=5"`
	Title              string `json:"title" validate:"max=100"`
	Content            string `json:"content" validate:"max=1000"`
	IsVerifiedPurchase bool   `json:"is_verified_purchase"`

	// Moderation
	Status           ReviewStatus `json:"status" gorm:"type:varchar(20);default:'PENDING'"`
	ModeratedBy      *uint        `json:"moderated_by"`
	ModeratedAt      *time.Time   `json:"moderated_at"`
	ModerationReason string       `json:"moderation_reason" validate:"max=500"`

	// Engagement
	HelpfulCount int `json:"helpful_count" gorm:"default:0"`

	// Relationships
	Images         []ReviewImage   `json:"images" gorm:"foreignKey:ProductReviewID"`
	SellerResponse *SellerResponse `json:"seller_response,omitempty" gorm:"foreignKey:ProductReviewID"`
	HelpfulVotes   []ReviewHelpful `json:"-" gorm:"foreignKey:ProductReviewID"`
}

// ReviewImage represents an image attached to a product review
type ReviewImage struct {
	gorm.Model
	ProductReviewID uint   `json:"product_review_id" gorm:"index"`
	URL             string `json:"url" validate:"required,url"`
	AltText         string `json:"alt_text" validate:"max=100"`
}

// SellerResponse represents a seller's response to a customer review
type SellerResponse struct {
	gorm.Model
	ProductReviewID uint   `json:"product_review_id" gorm:"uniqueIndex"` // One response per review
	UserID          uint   `json:"user_id" gorm:"index"`                 // Seller user ID
	User            User   `json:"user"`
	Content         string `json:"content" validate:"required,max=500"`
}

// ReviewHelpful tracks whether users found a review helpful or not
type ReviewHelpful struct {
	gorm.Model
	ProductReviewID uint `json:"product_review_id" gorm:"index"`
	UserID          uint `json:"user_id" gorm:"index"`
	IsHelpful       bool `json:"is_helpful"`
}

// ProductRating stores aggregated rating data for a product variant
type ProductRating struct {
	gorm.Model
	ProductVariantID uint    `json:"product_variant_id" gorm:"uniqueIndex"`
	AverageRating    float64 `json:"average_rating" gorm:"type:decimal(3,1);default:0.0"`
	TotalReviews     int     `json:"total_reviews" gorm:"default:0"`
	RatingBreakdown  string  `json:"rating_breakdown"` // JSON: {"1":0,"2":1,"3":2,"4":5,"5":10}
}

// TableName overrides the table name for ProductReview
func (ProductReview) TableName() string {
	return "product_reviews"
}

// TableName overrides the table name for ReviewImage
func (ReviewImage) TableName() string {
	return "review_images"
}

// TableName overrides the table name for SellerResponse
func (SellerResponse) TableName() string {
	return "seller_responses"
}

// TableName overrides the table name for ReviewHelpful
func (ReviewHelpful) TableName() string {
	return "review_helpful_votes"
}

// ReviewModerationLog tracks all moderation actions for audit purposes
type ReviewModerationLog struct {
	gorm.Model
	ReviewID    uint          `json:"review_id" gorm:"index"`
	Review      ProductReview `json:"review"`
	AdminID     uint          `json:"admin_id" gorm:"index"`
	Admin       User          `json:"admin"`
	OldStatus   ReviewStatus  `json:"old_status"`
	NewStatus   ReviewStatus  `json:"new_status"`
	Reason      string        `json:"reason" validate:"max=500"`
	ModeratedAt time.Time     `json:"moderated_at"`
}

// TableName overrides the table name for ProductRating
func (ProductRating) TableName() string {
	return "product_ratings"
}

// TableName overrides the table name for ReviewModerationLog
func (ReviewModerationLog) TableName() string {
	return "review_moderation_logs"
}

// BeforeCreate GORM hook to set default status for new reviews
func (r *ProductReview) BeforeCreate(tx *gorm.DB) error {
	if r.Status == "" {
		r.Status = ReviewStatusPending
	}
	return nil
}

// IsApproved returns true if the review is approved and visible to public
func (r *ProductReview) IsApproved() bool {
	return r.Status == ReviewStatusApproved
}

// IsPending returns true if the review is pending moderation
func (r *ProductReview) IsPending() bool {
	return r.Status == ReviewStatusPending
}

// IsRejected returns true if the review was rejected
func (r *ProductReview) IsRejected() bool {
	return r.Status == ReviewStatusRejected
}

// IsFlagged returns true if the review was flagged for review
func (r *ProductReview) IsFlagged() bool {
	return r.Status == ReviewStatusFlagged
}

// CanBeModifiedBy checks if a user can modify this review
func (r *ProductReview) CanBeModifiedBy(userID uint, userType UserType) bool {
	// Admin can modify any review
	if userType == Admin {
		return true
	}

	// User can modify their own review if it's not rejected
	if r.UserID == userID && r.Status != ReviewStatusRejected {
		return true
	}

	return false
}

// CanBeDeletedBy checks if a user can delete this review
func (r *ProductReview) CanBeDeletedBy(userID uint, userType UserType) bool {
	// Admin can delete any review
	if userType == Admin {
		return true
	}

	// User can delete their own review
	if r.UserID == userID {
		return true
	}

	return false
}

// HasSellerResponse returns true if the review has a seller response
func (r *ProductReview) HasSellerResponse() bool {
	return r.SellerResponse != nil
}

// GetReviewerName returns the reviewer's display name
func (r *ProductReview) GetReviewerName() string {
	if r.User.FirstName != "" && r.User.LastName != "" {
		return r.User.FirstName + " " + r.User.LastName
	}
	if r.User.FirstName != "" {
		return r.User.FirstName
	}
	return "Anonymous"
}
