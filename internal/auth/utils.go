package auth

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetUserIdFromContext(c *gin.Context) primitive.ObjectID {
	userIdStr := c.MustGet("userID").(string)
	userId, _ := primitive.ObjectIDFromHex(userIdStr)
	return userId
}
