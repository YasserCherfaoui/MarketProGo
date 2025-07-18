# Task 7: Review Helpfulness System - Documentation

## Overview
Implements the system for users to mark reviews as helpful or unhelpful, tracks votes per user per review, updates helpful counts, and prevents duplicate or self-votes.

## Endpoint

### Mark Review as Helpful/Unhelpful
- **Route:** `POST /api/v1/reviews/:id/helpful`
- **Access:** Authenticated users
- **Request Body:**
  ```json
  { "is_helpful": true } // or false
  ```
- **Behavior:**
  - User can mark a review as helpful or unhelpful
  - If the same vote is sent again, the vote is removed (toggle)
  - If a different vote is sent, the vote is updated
  - Users cannot vote on their own reviews
  - Only one vote per user per review is allowed
  - Helpful count is updated in the review record (only helpful votes count)

## Business Rules
- Only authenticated users can vote
- Users cannot vote on their own reviews
- Each user can only have one vote per review
- Voting the same way twice removes the vote (toggle)
- Changing vote updates the record and helpful count
- Only approved reviews can be voted on
- Helpful count reflects only helpful votes

## Data Structure

### ReviewHelpful Model
```go
type ReviewHelpful struct {
    gorm.Model
    ProductReviewID uint `json:"product_review_id" gorm:"index"`
    UserID          uint `json:"user_id" gorm:"index"`
    IsHelpful       bool `json:"is_helpful"`
}
```

### ProductReview Field
- `HelpfulCount int` — updated to reflect the number of helpful votes

## Example Request
```http
POST /api/v1/reviews/123/helpful
Content-Type: application/json
Authorization: Bearer <token>

{ "is_helpful": true }
```

## Example Response
```json
{
  "success": true,
  "message": "Vote recorded successfully",
  "data": {
    "review_id": 123,
    "is_helpful": true,
    "helpful_count": 5
  }
}
```

## Error Handling
- 400 Bad Request: Invalid review ID, invalid request body, self-vote
- 401 Unauthorized: User not authenticated
- 404 Not Found: Review not found or not approved
- 500 Internal Server Error: Database errors
- Consistent error response format via `utils/response`

## Testing
- Comprehensive unit tests for:
  - Marking as helpful/unhelpful
  - Toggling/removing vote
  - Changing vote
  - Preventing self-votes
  - Error scenarios (invalid ID, not found, unauthorized, invalid body)
- All tests pass and cover all business rules

## Security
- Only authenticated users can vote
- No duplicate or self-votes allowed
- SQL injection prevented via GORM parameterization

## Status
**Complete**

---

**Next Phase:** Task 8 - Seller Response Functionality 