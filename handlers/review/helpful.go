package review

import (
	"net/http"
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MarkReviewHelpfulRequest represents the request body for marking a review as helpful
type MarkReviewHelpfulRequest struct {
	IsHelpful bool `json:"is_helpful"`
}

// MarkReviewHelpful handles POST /api/v1/reviews/:id/helpful
// Allows authenticated users to mark a review as helpful or unhelpful
func (h *ReviewHandler) MarkReviewHelpful(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse review ID
	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REVIEW_ID", "Invalid review ID")
		return
	}

	// Parse request body
	var request MarkReviewHelpfulRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Check if review exists and is approved
	var review models.ProductReview
	err = h.db.Where("id = ? AND status = ?", reviewID, models.ReviewStatusApproved).First(&review).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateErrorResponse(c, http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found or not approved")
			return
		}
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve review")
		return
	}

	// Prevent users from voting on their own reviews
	if review.UserID == userID.(uint) {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "SELF_VOTE_NOT_ALLOWED", "You cannot vote on your own review")
		return
	}

	// Check if user has already voted on this review
	var existingVote models.ReviewHelpful
	err = h.db.Where("product_review_id = ? AND user_id = ?", reviewID, userID).First(&existingVote).Error

	if err == nil {
		// User has already voted
		if existingVote.IsHelpful == request.IsHelpful {
			// Same vote - remove the vote
			err = h.db.Delete(&existingVote).Error
			if err != nil {
				response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to remove vote")
				return
			}

			// Update review helpful count
			if request.IsHelpful {
				review.HelpfulCount--
			} else {
				// For unhelpful votes, we might want to track this differently
				// For now, we'll just remove the vote
			}

		} else {
			// Different vote - update the vote
			oldVote := existingVote.IsHelpful
			existingVote.IsHelpful = request.IsHelpful
			err = h.db.Save(&existingVote).Error
			if err != nil {
				response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update vote")
				return
			}

			// Update review helpful count
			if oldVote && !request.IsHelpful {
				// Changed from helpful to unhelpful
				review.HelpfulCount--
			} else if !oldVote && request.IsHelpful {
				// Changed from unhelpful to helpful
				review.HelpfulCount++
			}
		}

	} else if err == gorm.ErrRecordNotFound {
		// User hasn't voted yet - create new vote
		vote := models.ReviewHelpful{
			ProductReviewID: uint(reviewID),
			UserID:          userID.(uint),
			IsHelpful:       request.IsHelpful,
		}

		err = h.db.Create(&vote).Error
		if err != nil {
			response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create vote")
			return
		}

		// Update review helpful count
		if request.IsHelpful {
			review.HelpfulCount++
		}
	} else {
		// Database error
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to check existing vote")
		return
	}

	// Save the updated review helpful count
	err = h.db.Save(&review).Error
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update review helpful count")
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vote recorded successfully",
		"data": gin.H{
			"review_id":     reviewID,
			"is_helpful":    request.IsHelpful,
			"helpful_count": review.HelpfulCount,
		},
	})
}

// GetUserVoteStatus returns the current vote status for a user on a specific review
func (h *ReviewHandler) GetUserVoteStatus(userID, reviewID uint) (*models.ReviewHelpful, error) {
	var vote models.ReviewHelpful
	err := h.db.Where("product_review_id = ? AND user_id = ?", reviewID, userID).First(&vote).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No vote found
		}
		return nil, err
	}
	return &vote, nil
}

// UpdateReviewHelpfulCount updates the helpful count for a review based on all votes
func (h *ReviewHandler) UpdateReviewHelpfulCount(reviewID uint) error {
	var helpfulCount int64
	err := h.db.Model(&models.ReviewHelpful{}).
		Where("product_review_id = ? AND is_helpful = ?", reviewID, true).
		Count(&helpfulCount).Error
	if err != nil {
		return err
	}

	err = h.db.Model(&models.ProductReview{}).
		Where("id = ?", reviewID).
		Update("helpful_count", helpfulCount).Error
	return err
}
