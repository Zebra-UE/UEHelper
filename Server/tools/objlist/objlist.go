package objlist

import (
	"UEHelper/tools/factory"
	"regexp"
	"strings"
)

type ObjListItem struct {
	Class string
	Count int32
	NumKB float32
	MaxKB float32
}

type ObjList struct {
	Items []ObjListItem
}

type ProcessParam struct {
	lineRegexp *regexp.Regexp
	lastLine   bool
}

func processLine(line string, lineNum int32, params *ProcessParam, objList *ObjList) {

	if lineNum == 4 {

	} else if lineNum >= 5 {
		if len(line) == 0 {
			params.lastLine = true
		} else if params.lastLine {

		} else {
			match := params.lineRegexp.FindStringSubmatch(line)

			if len(match) > 0 {
				sizeArray := strings.Fields(match[3])
				if len(sizeArray) != 6 {

				}

			}
		}

	}
}

func Load(filePath string) *ObjList {
	var result ObjList
	var lineNum int32
	lineNum = 1
	var params ProcessParam
	params.lineRegexp = regexp.MustCompile(`(\w+)\s+(\d+)((?:\s+\d+\.\d{2}){6})`)
	//params.sizeRegexp = regexp.MustCompile(`(\s+\d+\.\d{2})`)

	factory.ReadLines(filePath, func(line string) {
		processLine(line, lineNum, &params, &result)
		lineNum += 1
	})
	return &result
}
