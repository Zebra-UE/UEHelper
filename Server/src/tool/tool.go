package tool

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Tool struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var registeredTools = []Tool{
	{ID: "1", Name: "分析BodySetup占用情况"},
	{ID: "2", Name: "工具2"},
	{ID: "3", Name: "工具3"},
}

func GetToolList(c *gin.Context) {
	c.JSON(http.StatusOK, registeredTools)
}

type RunToolPayload struct {
	ToolID   string `json:"toolId"`
	FileName string `json:"fileName"`
}

func RunTool(c *gin.Context) {
	var payload RunToolPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var result string
	if payload.ToolID == "1" {
		result = BodySetupTool{}.run(payload.FileName)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "任务已开始",
		"toolId":   payload.ToolID,
		"fileName": payload.FileName,
		"result":   result,
	})
}
