package main

import (
	loganalysis "UEHelper/tools/loganalysis"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main1() {
	r := gin.Default()
	r.POST("/log/analysis", func(c *gin.Context) {
		// var logFileInfo loganalysis.LogFileInfo
		// if err := c.ShouldBindJSON(&logFileInfo); err != nil {
		// 	result := loganalysis.Analysis(logFileInfo)
		// 	c.JSON(200, gin.H{"message": result})
		// }

	})

	r.Run("127.0.0.1:1122")
}

func main() {

	Path := "E:/Game/grgame/custom/release/S1Game/596701/Win64/S1Game_release_0.596701.596701.596701_596701_Development_Win64_OverSea/S1Game/Saved/Logs/S1Game.log"
	result := loganalysis.Analysis(Path)
	fmt.Print(result)
}

func main3() {

}
