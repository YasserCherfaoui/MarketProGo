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

// AdminMiddleware ensures the user has admin privileges
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First authenticate the user
		AuthMiddleware()(c)
		if c.IsAborted() {
			return
		}

		userType, exists := c.Get("user_type")
		if !exists {
			response.GenerateUnauthorizedResponse(c, "auth/middleware", "user type not found")
			c.Abort()
			return
		}

		if userType != models.Admin {
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
		// First authenticate the user
		AuthMiddleware()(c)
		if c.IsAborted() {
			return
		}

		userType, exists := c.Get("user_type")
		if !exists {
			response.GenerateUnauthorizedResponse(c, "auth/middleware", "user type not found")
			c.Abort()
			return
		}

		// Allow both Vendor and Admin to access seller routes
		if userType != models.Vendor && userType != models.Admin {
			response.GenerateForbiddenResponse(c, "auth/middleware", "seller access required")
			c.Abort()
			return
		}

		c.Next()
	}
}
