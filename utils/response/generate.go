package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // Use interface{} for flexibility
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func NewAPIError(code string, description string) *APIError {
	return &APIError{
		Code:        code,
		Description: description,
	}
}

func GenerateResponse(c *gin.Context, status int, message string, data interface{}, err *APIError) {
	c.JSON(status, APIResponse{
		Status:  message,
		Message: message,
		Data:    data,
		Error:   err,
	})
}

func GenerateSuccessResponse(c *gin.Context, message string, data interface{}) {
	GenerateResponse(c, http.StatusOK, message, data, nil)
}

func GenerateCreatedResponse(c *gin.Context, message string, data interface{}) {
	GenerateResponse(c, http.StatusCreated, message, data, nil)
}

func GenerateNoContentResponse(c *gin.Context, message string) {
	GenerateResponse(c, http.StatusNoContent, message, nil, nil)
}

func GenerateErrorResponse(c *gin.Context, status int, code string, description string) {
	err := NewAPIError(code, description)
	GenerateResponse(c, status, err.Description, nil, err)
}

func GenerateBadRequestResponse(c *gin.Context, code string, description string) {
	err := NewAPIError(code, description)
	GenerateResponse(c, http.StatusBadRequest, err.Description, nil, err)
}

func GenerateUnauthorizedResponse(c *gin.Context, code string, description string) {
	err := NewAPIError(code, description)
	GenerateResponse(c, http.StatusUnauthorized, err.Description, nil, err)
}

func GenerateForbiddenResponse(c *gin.Context, code string, description string) {
	err := NewAPIError(code, description)
	GenerateResponse(c, http.StatusForbidden, err.Description, nil, err)
}

func GenerateNotFoundResponse(c *gin.Context, code string, description string) {
	err := NewAPIError(code, description)
	GenerateResponse(c, http.StatusNotFound, err.Description, nil, err)
}

func GenerateInternalServerErrorResponse(c *gin.Context, code string, description string) {
	err := NewAPIError(code, description)
	GenerateResponse(c, http.StatusInternalServerError, err.Description, nil, err)
}
