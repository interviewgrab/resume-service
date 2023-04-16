package auth

import (
	"net/http"
	"resume-service/internal/database"

	"github.com/gin-gonic/gin"
)

func EmailVerified(userStore *database.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := GetUserIdFromContext(c)

		user, err := userStore.GetUser(c, userId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "email_not_verified"})
			return
		}

		if user.EmailVerified == false {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "email_not_verified"})
			return
		}
		c.Next()
	}
}
