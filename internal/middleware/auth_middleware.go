package middleware

import (
	"auth-service/internal/domain"
	"auth-service/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(service *service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := GetToken(c)
		claims, err := service.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, domain.NewErrorResponse("Invalid token", err.Error()))
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}

func GetToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}
