package main

import (
	loganalysis "UEHelper/tools/loganalysis"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main1() {
	r := gin.Default()
	r.POST("/log/analysis", func(c *gin.Context) {
		var logFileInfo loganalysis.LogFileInfo
		if err := c.ShouldBindJSON(&logFileInfo); err != nil {
			result := loganalysis.Analysis(logFileInfo)
			c.JSON(200, gin.H{"message": result})
		}

	})

	r.Run("127.0.0.1:1122")
}

func main2() {
	var logFileInfo loganalysis.LogFileInfo
	logFileInfo.Path = "E:/Game/S1Game_release_0.578987.578987.578987_578987_Development_Win64_OverSea/S1Game/Saved/Logs/S1Game.log"
	result := loganalysis.Analysis(logFileInfo)
	fmt.Print(result)
}

func main() {

}
