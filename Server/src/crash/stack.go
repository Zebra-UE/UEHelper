package crash

import (
	"regexp"
	"strconv"
	"strings"
)

type FStackRow struct {
	Index    int
	Module   string
	Pointer  string
	FilePath string
	LineNum  int
	FuncName string
	Param    string
}

type FStackArray struct {
	StackRows []FStackRow
}

func FormatStack(stackStr string) FStackArray {
	var stackArray FStackArray
	lines := strings.Split(stackStr, "\n")

	for i, line := range lines {
		if len(line) > 0 {
			row := parseStackLine(line)
			row.Index = i
			stackArray.StackRows = append(stackArray.StackRows, row)
		}
	}
	return stackArray
}

func parseStackLine(line string) FStackRow {
	// S1Game.exe	pc 0x17e3610	static FMallocBinnedCommonUtils::TrimThreadFreeBlockLists(FMallocBinned2 &, TMallocBinnedCommon::FPerThreadFreeBlockLists *) (D:\projects\release\UE5EA\Engine\Source\Runtime\Core\Public\HAL/MallocBinnedCommonUtils.h:112)[amd64:Windows NT:8CE0C996D3B54EB58FE8A10A7619CDD01]
	re := regexp.MustCompile(`(?P<module>.*?)\s+pc\s+(?P<pointer>0x[0-9a-fA-F]+)\s+(?P<func>.*?)\s+\((?P<path>.*?):(?P<line>\d+)\)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) == 0 {
		// Fallback for lines without file path and line number
		re = regexp.MustCompile(`(?P<module>.*?)\s+pc\s+(?P<pointer>0x[0-9a-fA-F]+)\s+(?P<func>.*)`)
		matches = re.FindStringSubmatch(line)
		if len(matches) == 0 {
			return FStackRow{FuncName: line}
		}

		funcName := strings.TrimSpace(matches[3])
		funcAndParam := strings.SplitN(funcName, "(", 2)
		param := ""
		if len(funcAndParam) > 1 {
			param = "(" + funcAndParam[1]
		}

		return FStackRow{
			Module:   matches[1],
			Pointer:  matches[2],
			FuncName: funcAndParam[0],
			Param:    param,
		}
	}

	funcName := strings.TrimSpace(matches[3])
	funcAndParam := strings.SplitN(funcName, "(", 2)
	param := ""
	if len(funcAndParam) > 1 {
		param = "(" + funcAndParam[1]
	}

	lineNum, _ := strconv.Atoi(matches[5])
	return FStackRow{
		Module:   matches[1],
		Pointer:  matches[2],
		FuncName: funcAndParam[0],
		Param:    param,
		FilePath: matches[4],
		LineNum:  lineNum,
	}
}

func ReBucket(stackArray *FStackArray) float {
	startStack := -1
	endStack := -1
	for i, row := range stackArray.StackRows {
		if startStack == -1 {
			if row.Module == "S1Game.exe" {
				startStack = i
			}
		} else if endStack == -1 {
			if row.Module == "S1Game.exe" {
				if row.FuncName == "GuardedMain" {
					endStack = i
				}
			}
		}
	}
	if endStack == -1 {
		endStack = len(stackArray.StackRows)
	}
	if startStack == -1 {
		return 0
	}

}
