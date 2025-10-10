package loganalysis

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type LogFileInfo struct {
	Path string `json:"path"`
}

type LogBasicInfo struct {
	Path      string `json:"path"`
	StartTime string `json:"start_time"`
}

type LogAnalysisResult struct {
	BasicInfo LogBasicInfo `json:"basic_info"`
}

func processLine(index int, line string, result *LogAnalysisResult) {
	if index == 0 {
		// Log file open, 09/28/25 16:14:45
		result.BasicInfo.StartTime = line[15:]
		return
	}
	if strings.HasPrefix(line, "LogPakFile:") {

	}
}

func Analysis(c *gin.Context) string {
	var logFileInfo LogFileInfo
	if err := c.ShouldBindJSON(&logFileInfo); err != nil {
		var result LogAnalysisResult

		result.BasicInfo.Path = logFileInfo.Path

		file, err := os.Open(logFileInfo.Path)
		if err != nil {
			return ""
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		buf := make([]byte, 0, 64*1024) // 自定义缓冲区大小（64KB）
		scanner.Buffer(buf, 1024*1024)  // 设置最大行长度（1MB）
		lineNum := 0
		for scanner.Scan() {
			line := scanner.Text()
			processLine(lineNum, line)
			lineNum += 1
		}
		if err := scanner.Err(); err != nil {
			return ""
		}
		jsonData, err := json.Marshal(result)
		if err != nil {
			return ""
		}
		return string(jsonData)
	}

	return ""
}
