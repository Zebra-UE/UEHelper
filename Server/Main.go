package main

import (
	"UEHelper/src/loganalysis"
	"UEHelper/src/tool"

	"github.com/gin-gonic/gin"
)

const baseDir = "E:/Game/grgame/custom"

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	//
	//fmt.Printf(os.Getwd())

	r.GET("/loganalysis", loganalysis.View)
	r.POST("/api/loganalysis/run", loganalysis.AnalyzeAPI)

	r.GET("/tool", func(c *gin.Context) {
		c.HTML(200, "tool.html", nil)
	})
	r.GET("/api/tool/list", tool.GetToolList)
	r.POST("/api/tool/run", tool.RunTool)

	r.GET("/tool/run", func(c *gin.Context) {
		c.HTML(200, "tool_run.html", nil)
	})

	r.GET("/", func(ctx *gin.Context) {
		ctx.File("./templates/index.html")
	})

	r.Run("127.0.0.1:1122")
}
