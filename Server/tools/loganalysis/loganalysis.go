package loganalysis

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type LogFileInfo struct {
	Path string `json:"path"`
}

type PakInfo struct {
	Name       string `json:"name"`
	MountPoint string `json:"mount_point"`
	Success    bool   `json:"success"`
}

type LogBasicInfo struct {
	Path        string    `json:"path"`
	StartTime   string    `json:"start_time"`
	PakInfos    []PakInfo `json:"pak_infos"`
	Commandline string    `json:"commandline"`
}

type LogAnalysisResult struct {
	BasicInfo LogBasicInfo `json:"basic_info"`
}

type BasicInfoRegexpItem struct {
	prefix  string
	times   int
	re      *regexp.Regexp
	process func([]string, *LogAnalysisResult)
}

type BasicInfoRegexp struct {
	items []BasicInfoRegexpItem
}

func (me *BasicInfoRegexp) processLine(index int, line string, result *LogAnalysisResult) {
	for _, item := range me.items {
		if item.times == 0 {
			continue
		} else {
			if strings.HasPrefix(line, item.prefix) {
				ret := item.re.FindStringSubmatch(line)
				if len(ret) > 0 {
					item.process(ret, result)
					if item.times > 0 {
						item.times -= 1
					}
					return
				}
			}
		}
	}

}

func process_start_time(ret []string, result *LogAnalysisResult) {
	result.BasicInfo.StartTime = ret[1]
}

func process_found_pak(ret []string, result *LogAnalysisResult) {
	result.BasicInfo.PakInfos = append(result.BasicInfo.PakInfos, PakInfo{
		Name:    ret[1],
		Success: false,
	})
}
func process_mounted_pak(ret []string, result *LogAnalysisResult) {
	pak_name := ret[1]

	for i := 0; i < len(result.BasicInfo.PakInfos); i++ {
		if result.BasicInfo.PakInfos[i].Name == pak_name {
			result.BasicInfo.PakInfos[i].Success = true
			result.BasicInfo.PakInfos[i].MountPoint = ret[2]
			break
		}
	}
}

func process_commandline(ret []string, result *LogAnalysisResult) {
	result.BasicInfo.Commandline = ret[1]
}

func (me *BasicInfoRegexp) construct() {
	me.items = append(me.items, BasicInfoRegexpItem{
		prefix:  "Log file open, ",
		times:   1,
		re:      regexp.MustCompile(`^Log file open, (.*)$`),
		process: process_start_time,
	}, BasicInfoRegexpItem{
		prefix:  "LogPakFile:",
		times:   -1,
		re:      regexp.MustCompile(`^LogPakFile: Display: Found Pak file (.*) attempting to mount.$`),
		process: process_found_pak,
	}, BasicInfoRegexpItem{
		prefix:  "LogPakFile:",
		times:   -1,
		re:      regexp.MustCompile(`^LogPakFile: Display: Mounted Pak file '(.*)', mount point: '(.*)'$`),
		process: process_mounted_pak,
	}, BasicInfoRegexpItem{
		prefix:  "LogCsvProfiler:",
		times:   1,
		re:      regexp.MustCompile(`^LogCsvProfiler: Display: Metadata set : commandline="(.*)"$`),
		process: process_commandline,
	})
}

type LogLineRegexpItem struct {
	prefix       string
	begin_regexp *regexp.Regexp
	end_regexp   *regexp.Regexp
	process      func(int, string, string, string, *LogAnalysisResult)
}
type LogLineRegexp struct {
	index int
	items []LogLineRegexpItem
}

func (me *LogLineRegexp) processLine(line_num int, time string, frame string, line string, result *LogAnalysisResult) {
	if me.index == -1 {
		for i, item := range me.items {
			if strings.HasPrefix(line, item.prefix) {
				ret := item.begin_regexp.FindStringSubmatch(line)
				if len(ret) > 0 {
					if item.end_regexp != nil {
						me.index = i
					} else {
						item.process(line_num, time, frame, line, result)
					}
					return
				}
			}
		}
	} else {
		item := me.items[me.index]
		if item.end_regexp != nil {
			ret := item.end_regexp.FindStringSubmatch(line)
			if len(ret) > 0 {
				me.index = -1
				item.process(line_num, time, frame, line, result)
				return
			} else {
				item.process(line_num, time, frame, line, result)
			}

		}
	}
}

func process_reference_chain(line_num int, time string, frame string, line string, result *LogAnalysisResult) {

}

func Analysis(logFileInfo LogFileInfo) string {

	var result LogAnalysisResult

	result.BasicInfo.Path = logFileInfo.Path

	file, err := os.Open(logFileInfo.Path)
	if err != nil {
		return ""
	}
	defer file.Close()

	bom := []byte{0xEF, 0xBB, 0xBF}
	buffer := make([]byte, 3)
	_, err = file.Read(buffer)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return ""
	}

	if !bytes.Equal(buffer, bom) {
		_, err = file.Seek(0, 0)
		if err != nil {
			fmt.Println("Error seeking file:", err)
			return ""
		}
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024) // 自定义缓冲区大小（64KB）
	scanner.Buffer(buf, 1024*1024)  // 设置最大行长度（1MB）
	lineNum := 0
	var basic_info_regexp BasicInfoRegexp
	basic_info_regexp.construct()
	log_regexp := regexp.MustCompile(`^\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{3})\]\[\s*(\d+)\](.*)$`)

	var logline_factory LogLineRegexp
	logline_factory.index = -1
	logline_factory.items = []LogLineRegexpItem{
		{
			prefix:       "LogReferenceChain:",
			begin_regexp: regexp.MustCompile(`^LogReferenceChain: Display: InitialGather memory usage:`),
			end_regexp:   regexp.MustCompile((`LogReferenceChain: Display: Post-search memory usage:`)),
			process:      process_reference_chain,
		},
		{
			prefix: "LogLoad: LoadMap:",
			begin_regexp: 
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "[") {
			basic_info_regexp.processLine(lineNum, line, &result)
		} else {
			ret := log_regexp.FindStringSubmatch(line)
			if ret != nil {
				logline_factory.processLine(lineNum, ret[1], ret[2], ret[3], &result)
			}
		}
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
