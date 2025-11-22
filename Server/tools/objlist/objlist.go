package objlist

import (
	"UEHelper/tools/factory"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ClassListItem struct {
	Count int32
}

type ClassList struct {
	Items map[string]ClassListItem
	Total int32
}

type ProcessParam struct {
	lineRegexp *regexp.Regexp
	lastLine   bool
}

func processLine(line string, lineNum int32, params *ProcessParam, classList *ClassList) {

	if lineNum <= 4 {
		if lineNum == 1 {

		}

	} else if lineNum >= 5 {
		if len(line) == 0 {
			params.lastLine = true
		} else if params.lastLine {

		} else {
			match := params.lineRegexp.FindStringSubmatch(line)
			if len(match) > 0 {
				var item ClassListItem
				className := match[1]
				count, _ := strconv.Atoi(match[2])
				sizeArray := strings.Fields(match[3])
				if len(sizeArray) != 6 {

				}
				item.Count = int32(count)
				classList.Items[className] = item

			}
		}
	}
}

func Load(filePath string) *ClassList {
	var classList ClassList
	classList.Items = make(map[string]ClassListItem)
	var lineNum int32
	lineNum = 1
	var params ProcessParam
	params.lineRegexp = regexp.MustCompile(`(\w+)\s+(\d+)((?:\s+\d+\.\d{2}){6})`)
	//params.sizeRegexp = regexp.MustCompile(`(\s+\d+\.\d{2})`)

	result := factory.ReadLines(filePath, func(line string) bool {
		if lineNum == 1 {
			if strings.Contains(line, "class=") {
				return false
			}
		}
		processLine(line, lineNum, &params, &classList)
		lineNum += 1

		return true
	})
	if result == false {
		return nil
	}
	return &classList
}

func CompareClassList(filepaths ...string) {

	classListArray := make([]*ClassList, 0)
	for _, req := range filepaths {
		classList := Load(req)
		if classList != nil {
			classListArray = append(classListArray, classList)
		}
	}
	if len(classListArray) < 2 {
		return
	}
	//count := len(classListArray)-1
	compareResult := make([]map[string]int32, len(classListArray)-1)
	result := make(map[string]int32)
	for i := 1; i < len(classListArray); i++ {
		prevObjList := classListArray[i-1]
		currObjList := classListArray[i]
		diffMap := make(map[string]int32)
		for className, currItem := range currObjList.Items {
			prevItem, exists := prevObjList.Items[className]
			if exists {
				diff := currItem.Count - prevItem.Count
				if diff > 0 {
					diffMap[className] = diff
					result[className] = 0
				}
			} else {
				diffMap[className] = currItem.Count
				result[className] = 0
			}
		}
		compareResult[i-1] = diffMap
	}

	for className := range result {
		count := 0
		for i := 0; i < len(compareResult); i++ {
			_, exists := compareResult[i][className]
			if exists {
				count += 1
			}
		}
		if count >= len(compareResult) {
			fmt.Print(className + "\n")
		}
	}

}

func CompareObjectList(filepaths ...string) {

}
