package loganalysis

import (
	"UEHelper/tools/factory"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Stage interface {
	construct()
	isMatch(line string) bool
	process(timestamp string, framecount string, line string) bool
	Printf()
}

type PakInfo struct {
	source string
	mount  string
}

type ThreadInfo struct {
}

type BasicInfo struct {
	start_time string
	pak_info   []PakInfo
}

func (me *BasicInfo) isMatch(line string) bool {
	return false
}

func (me *BasicInfo) process(timestamp string, framecount string, line string) bool {
	return false

}

type ClassSizeTuple struct {
	Count int `json:"count"`
	Size  int `json:"size"`
}

type ClassSize struct {
	regexp     *regexp.Regexp
	class_size map[string]ClassSizeTuple
}

func (me *ClassSize) construct() {
	me.regexp = regexp.MustCompile(`LogS1GameAssetSubsystem:\s+Display:\s+(.+)\s+(\d+)\s+(\d+)`)
	me.class_size = make(map[string]ClassSizeTuple)
}

func (me ClassSize) MarshalJSON() ([]byte, error) {

	return json.Marshal(struct {
		Type      string                    `json:"type"`
		ClassSize map[string]ClassSizeTuple `json:"class_size"`
	}{
		Type:      "ClassSize",
		ClassSize: me.class_size,
	})
}

func (me *ClassSize) Printf() {
	for name, tuple := range me.class_size {
		fmt.Print(name, ",", tuple.Count, ",", tuple.Size, "\n")
	}
}

func (me *ClassSize) isMatch(line string) bool {
	return strings.HasPrefix(line, "LogS1GameAssetSubsystem")
}

func (me *ClassSize) process(timestamp string, framecount string, line string) bool {

	if me.regexp == nil {
		me.construct()
	}
	match := me.regexp.FindStringSubmatch(line)
	if match != nil {
		name := match[1]
		count, _ := strconv.Atoi(match[2])
		size, _ := strconv.Atoi(match[3])

		_, ok := me.class_size[name]
		if !ok {
			me.class_size[name] = ClassSizeTuple{Count: count, Size: size}
			return true
		}

	}
	return false

}

func NewClassSize() *ClassSize {
	cs := &ClassSize{}
	cs.construct() // 调用初始化逻辑
	return cs
}

func Analysis(path string) string {
	var stages []Stage = []Stage{NewClassSize()}
	timestamp_regexp := regexp.MustCompile(`^\[(\d{4}\.\d{2}\.\d{2}\-\d{2}\.\d{2}\.\d{2}:\d{3})\]\[\s{0,3}(\d{1,3})\](.*)$`)
	is_basic_info := true
	factory.ReadLines(path, func(line string) {
		if is_basic_info {
			if strings.HasPrefix(line, "[") {
				is_basic_info = false
			}
		}
		if is_basic_info {

		} else {
			match := timestamp_regexp.FindStringSubmatch(line)
			if match != nil {
				for _, stage := range stages {
					if stage.isMatch(match[3]) {
						if stage.process(match[1], match[2], match[3]) {
							break
						}
					}
				}
			}

		}

	})

	for _, stage := range stages {
		stage.Printf()
	}

	return ""
}
