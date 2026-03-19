package web

import (
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

type FLaunchManager struct {
}

type FLaunch struct {
}

func View(ctx *gin.Context) {

	ctx.HTML(200, "launch.html", nil)
}
func (self *FLaunch) List(ctx *gin.Context) {
	var req struct {
		Branch string `json:"branch"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	var rsp struct {
		Changelist []string `json:"files"`
	}
	root := "E:/Game/grgame/custom"

	branchPath := root
	if req.Branch == "" {

		return
	}
	switch req.Branch {
	case "Trunk":
		branchPath = path.Join(branchPath, "trunk")
	case "Release":
		branchPath = path.Join(branchPath, "release")
	default:
		return
	}
	branchPath = path.Join(branchPath, "S1Game")
	files, err := os.ReadDir(branchPath)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to read directory"})
		return
	}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		rsp.Changelist = append(rsp.Changelist, file.Name())
	}
	ctx.JSON(200, rsp)
}
