package main

import (
	loganalysis "UEHelper/tools"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/log/analysis", func(c *gin.Context) {
		result := loganalysis.Analysis(c)
		c.JSON(200, gin.H{"message": result})
	})

	r.Run("127.0.0.1:1122")
}
