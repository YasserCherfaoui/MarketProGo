package middlewares

import (
	"strings"

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
