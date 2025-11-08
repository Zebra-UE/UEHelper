package main

import (
	"UEHelper/tools/objlist"
	"UEHelper/tools/pakview"

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

	pakview.Load("E:/Game/grgame/custom/release/S1Game/633821/Win64/S1Game/Content/Paks/pakchunk0-Windows.pak")
}

func main3() {
	objlist.Load("E:/Game/grgame/custom/release/S1Game/628706/Win64/S1Game_release_0.628706.628706.628706_628706_Development_Win64_OverSea/S1Game/Saved/Profiling/NordLand-Windows-628706/Pid29160_ObjectStatistics_1.dumpobj")
}
