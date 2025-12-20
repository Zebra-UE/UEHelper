package main

import (
	"UEHelper/tools/crash"
	"UEHelper/tools/loganalysis"
	"UEHelper/tools/objlist"
	"UEHelper/tools/pakview"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	//
	//fmt.Printf(os.Getwd())
	r.GET("/api/crash/list", func(c *gin.Context) {
		//absPath := "C:/Users/36038/Downloads"
		//c.JSON(crash.List(absPath))
	})
	r.GET("api/crash/:name", func(c *gin.Context) {
		name := c.Param("name")
		absPath := filepath.Join("C:/Users/36038/Downloads", name)
		extList := []string{".dmp.gz", ".7z"}
		for _, ext := range extList {
			absFilePath := absPath + ext
			_, err := os.Stat(absFilePath)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
			}
			c.String(200, "%s", gin.H{"message": crash.Run(absFilePath)})
			return
		}

	})

	r.GET("/crash", func(ctx *gin.Context) {
		ctx.File("./templates/crash.html")
	})

	r.GET("/loganalysis", func(ctx *gin.Context) {
		ctx.File("./templates/loganalysis.html")
	})

	r.POST("/api/loganalysis/", func(ctx *gin.Context) {
		var req struct {
			Path string `json:"path"`
		}

		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		if req.Path == "" {
			ctx.JSON(400, gin.H{"error": "Path is required"})
			return
		}

		// 调用日志分析函数
		result := loganalysis.Run(req.Path)

		ctx.JSON(200, result)
	})

	r.GET("/", func(ctx *gin.Context) {
		ctx.File("./templates/index.html")
	})

	r.Run("127.0.0.1:1122")
}

func main2() {

	pakview.Load("E:/Game/grgame/custom/release/S1Game/633821/Win64/S1Game/Content/Paks/pakchunk0-Windows.pak")
}

func main1() {
	baseDir := "E:/Game/grgame/custom"
	branch := "release"
	changelist := "678906"
	profilePath := [2]string{
		filepath.Join(baseDir, branch, "S1Game", changelist, "Win64", "S1Game", "Saved", "Profiling"),
		filepath.Join(baseDir, branch, "S1Game", changelist, "Win64", "xxx", "S1Game", "Saved", "Profiling"),
	}
	entries, err := os.ReadDir(profilePath[0])
	if err != nil {

	}
	var scanPath string
	if len(entries) > 0 {
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() == "MemReports" {
				continue
			}
			arr := strings.Split(entry.Name(), "-")
			if len(arr) < 3 {
				continue
			}
			if arr[1] == "Windows" {
				scanPath = filepath.Join(profilePath[0], entry.Name())
			}
		}
		if len(scanPath) == 0 {
			return
		}
	}
	type FindFileResultItem struct {
		pid   string
		order int
		path  string
	}
	findResult := make(map[string][]FindFileResultItem, 0)

	entries, _ = os.ReadDir(scanPath)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		if strings.HasPrefix(filename, "Pid") && strings.HasSuffix(filename, ".dumpobj") {
			arr := strings.Split(filename[:len(filename)-len(".dumpobj")], "_")
			if len(arr) == 3 {
				pid := arr[0][3:]
				order, _ := strconv.Atoi(arr[1])
				_, ok := findResult[pid]
				if !ok {
					item := make([]FindFileResultItem, 0)
					findResult[pid] = item
				}
				findResult[pid] = append(findResult[pid], FindFileResultItem{pid: pid, order: order, path: filepath.Join(scanPath, filename)})
			}
		}
	}

	for _, arr := range findResult {
		var requestPaths []string
		sort.Slice(arr, func(i, j int) bool {
			return arr[i].order < arr[j].order
		})
		for _, item := range arr {
			requestPaths = append(requestPaths, item.path)
		}
		objlist.CompareClassList(requestPaths...)
	}
}

func main4() {
	crash.Run("E:/UECC-Windows-27F5385844B8466AA3B356B5F49603D7_0000.7z")
}
