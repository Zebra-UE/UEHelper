package loganalysis

import (
	"UEHelper/tools/factory"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ParseUETimestampWithMs(timeStr string) int64 {
	timeStr = strings.Trim(timeStr, "[]")

	// 分离时间和毫秒部分
	parts := strings.Split(timeStr, ":")
	if len(parts) < 2 {
		return 0
	}

	// 时间部分: 2025.12.06-16.11.21
	timePart := parts[0]
	// 毫秒部分: 303
	msPart := parts[1]

	// 使用原始格式直接Parse: 2006.01.02-15.04.05
	t, err := time.Parse("2006.01.02-15.04.05", timePart)
	if err != nil {
		return 0
	}

	ms := 0
	if val, err := strconv.Atoi(msPart); err == nil {
		ms = val
	}

	return t.Unix()*1000 + int64(ms)
}

// 判断行是否符合日志格式，返回(是否匹配, 时间戳, 框号, 剩余字符串)
func IsValidLogLine(line string) (bool, string, int, string) {
	// 使用捕获组提取时间戳、框号和剩余部分
	pattern := `^\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{3,4})\]\[\s*(\d+)\](.*)$`

	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(line)

	if matches == nil || len(matches) < 4 {
		return false, "", 0, ""
	}

	timestamp := matches[1]                   // 第一个捕获组：时间戳
	frameNum, err := strconv.Atoi(matches[2]) // 第二个捕获组：框号转int
	if err != nil {
		return false, "", 0, ""
	}
	remaining := matches[3] // 第三个捕获组：剩余字符串

	return true, timestamp, frameNum, remaining
}

type LogLine struct {
	Timestamp string
	Frame     int
}

func (me *LogLine) Init(timestamp string, frame int) {
	me.Timestamp = timestamp
	me.Frame = frame

}

type LoadMap struct {
	LogLine
	MapName string
}

type FPhysicsMemory struct {
	LogLine
	Prefix     string           `json:"prefix"`
	MapName    string           `json:"map_name"`
	FrameCount int64            `json:"frame_count"`
	TotalSize  int64            `json:"total_size"`
	SizeMap    map[string]int64 `json:"size_map"`
}

type ILogLine interface {
	Init(timestamp string, frame int)
}
type ILogProcessor interface {
	Format(line string) ILogLine
	Output() interface{}
}
type FLogProcessor struct {
	processor map[string]ILogProcessor
}

type FPhysicsMemoryProcessor struct {
	PhysicsMemory []*FPhysicsMemory
}

func (me *FPhysicsMemoryProcessor) Format(line string) ILogLine {
	const prefix = "LogCollisionStriper: Display: [UCollisionStripperLoader::PrintMemoryInfo]"
	if strings.HasPrefix(line, prefix) {
		// 删除前缀
		line = strings.TrimPrefix(line, prefix)
		line = strings.TrimSpace(line)

		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			return nil
		}
		parts = strings.Split(line, "|")
		if len(parts) < 2 {
			return nil
		}
		//Initialize|MapName:HallMain|Physics:449.607 MB (471446823)|ChaosTrimesh:256.062 MB (268500596)

		result := FPhysicsMemory{}
		result.Prefix = strings.TrimSpace(parts[0])
		result.SizeMap = make(map[string]int64)

		extract_size := func(value string) int64 {
			pattern := `\((\d+)\)`
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(value)
			if len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					return val
				}
			}
			return 0
		}

		for i := 1; i < len(parts); i++ {
			subParts := strings.SplitN(parts[i], ":", 2)
			if len(subParts) < 2 {
				continue
			}
			key := strings.TrimSpace(subParts[0])
			value := strings.TrimSpace(subParts[1])
			switch key {
			case "MapName":
				result.MapName = value
			case "Frame":
				result.FrameCount, _ = strconv.ParseInt(value, 10, 64)
			case "Physics":
				result.TotalSize = extract_size(value)
			default:
				result.SizeMap[key] = extract_size(value)
			}
		}

		me.PhysicsMemory = append(me.PhysicsMemory, &result)
		return &result
	}
	return nil

}

func (me *FPhysicsMemoryProcessor) Output() interface{} {
	return me.PhysicsMemory
}

func (me *FLogProcessor) ProcessLine(timestamp string, frame int, line string) {
	for _, processor := range me.processor {
		if formatLine := processor.Format(line); formatLine != nil {
			formatLine.Init(timestamp, frame)
			return
		}
	}
}

func (me *FLogProcessor) Output() map[string]interface{} {

	result := make(map[string]interface{})
	for key, processor := range me.processor {
		result[key] = processor.Output()
	}

	return result
}

func Run(filepath string) map[string]interface{} {
	lineNum := 0

	logProcessor := FLogProcessor{}
	logProcessor.processor = make(map[string]ILogProcessor)
	logProcessor.processor["PhysicsMemory"] = &FPhysicsMemoryProcessor{}

	factory.ReadLines(filepath, func(rawLine string) bool {
		lineNum++
		if lineNum == 1 {
			if strings.HasPrefix(rawLine, "Log file open") {

			}
		} else {
			if ok, timestamp, frame, line := IsValidLogLine(rawLine); ok {
				logProcessor.ProcessLine(timestamp, frame, line)
			} else {

			}
		}
		return true
	})

	return logProcessor.Output()
}
