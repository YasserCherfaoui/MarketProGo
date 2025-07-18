package middlewares

import (
	"strings"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/auth"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			response.GenerateUnauthorizedResponse(c, "auth/middleware", "token is required")
			c.Abort()
			return
		}
		// remove the Bearer prefix
		token = strings.TrimPrefix(token, "Bearer ")
		claims, err := auth.ValidateToken(token)
		if err != nil {
			response.GenerateUnauthorizedResponse(c, "auth/middleware", "token is invalid")
			c.Abort()
			return
		}

		c.Set("user", claims)
		c.Set("user_id", claims.UserID)
		c.Set("user_type", claims.UserType)

		c.Next()
	}
}

// OptionalAuthMiddleware allows routes to work with or without authentication
// If a valid token is provided, user context is set; if not, the request continues without user context
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			// No token provided, continue without user context
			c.Next()
			return
		}

		// Remove the Bearer prefix
		token = strings.TrimPrefix(token, "Bearer ")
		claims, err := auth.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without user context
			c.Next()
			return
		}

		// Valid token, set user context
		c.Set("user", claims)
		c.Set("user_id", claims.UserID)
		c.Set("user_type", claims.UserType)

		c.Next()
	}
}

// AdminMiddleware ensures the user has admin privileges
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		token := c.GetHeader("Authorization")
		if token == "" {
			response.GenerateUnauthorizedResponse(c, "auth/middleware", "token is required")
			c.Abort()
			return
		}

		// Remove the Bearer prefix
		token = strings.TrimPrefix(token, "Bearer ")
		claims, err := auth.ValidateToken(token)
		if err != nil {
			response.GenerateUnauthorizedResponse(c, "auth/middleware", "token is invalid")
			c.Abort()
			return
		}

		// Set user context
		c.Set("user", claims)
		c.Set("user_id", claims.UserID)
		c.Set("user_type", claims.UserType)

		// Check if user is admin
		if claims.UserType != models.Admin {
			response.GenerateForbiddenResponse(c, "auth/middleware", "admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}

// SellerMiddleware ensures the user has seller privileges
func SellerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		token := c.GetHeader("Authorization")
		if token == "" {
			response.GenerateUnauthorizedResponse(c, "auth/middleware", "token is required")
			c.Abort()
			return
		}

		// Remove the Bearer prefix
		token = strings.TrimPrefix(token, "Bearer ")
		claims, err := auth.ValidateToken(token)
		if err != nil {
			response.GenerateUnauthorizedResponse(c, "auth/middleware", "token is invalid")
			c.Abort()
			return
		}

		// Set user context
		c.Set("user", claims)
		c.Set("user_id", claims.UserID)
		c.Set("user_type", claims.UserType)

		// Allow both Vendor and Admin to access seller routes
		if claims.UserType != models.Vendor && claims.UserType != models.Admin {
			response.GenerateForbiddenResponse(c, "auth/middleware", "seller access required")
			c.Abort()
			return
		}

		c.Next()
	}
}
