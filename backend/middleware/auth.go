package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ileego/go_browser_capture/backend/domain"
	"github.com/ileego/go_browser_capture/backend/infrastructure"
)

func AuthMiddleware(userRepo *infrastructure.PostgresUserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未提供认证令牌"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "无效的认证令牌格式"})
			c.Abort()
			return
		}

		claims, err := ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "令牌已过期或无效"})
			c.Abort()
			return
		}

		user, err := userRepo.FindByID(c, claims.UserID)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户不存在"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("claims", claims)
		c.Next()
	}
}

func PermissionMiddleware(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未登录"})
			c.Abort()
			return
		}

		if u, ok := user.(*domain.User); ok {
			if !u.HasPermission(permission) {
				c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "没有访问权限"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}