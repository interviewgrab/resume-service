package utils

import "github.com/gin-gonic/gin"

func GinError(err error) gin.H {
	return gin.H{"error": err.Error()}
}
