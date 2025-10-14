package memreport

import (
	"UEHelper/tools/factory"
	"encoding/json"
	"strconv"
	"strings"
)

type ClassItem struct {
	class_name string
	count      int
}
type TotalObject struct {
	count int
}
type ObjList struct {
	class_name  string
	commandline string
	items       map[string]ClassItem
	total       TotalObject
}

func read_objlist_from_file(path string) ObjList {
	var objlist ObjList
	line_num := 1
	var key_index [2]int = [2]int{-1, -1}
	var is_obj_list bool = false

	to_num := func(s string) int {
		num, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		return num
	}

	factory.ReadLines(path, func(line string) {
		line = strings.TrimSpace(line)
		if line_num == 1 {
			if strings.HasPrefix(line, "Obj List:") {
				commandline := line[10:]
				objlist.commandline = commandline
				if strings.HasPrefix(commandline, "class=") {
					objlist.class_name = commandline[6:]
				}
			}
		} else if line_num == 4 {
			cols := strings.Fields(line)
			for i := 0; i < len(cols); i++ {
				if cols[i] == "Class" {
					key_index[0] = i
				} else if cols[i] == "Count" {
					key_index[1] = i
				}
			}
			is_obj_list = true
		} else if line_num > 4 {
			if len(line) == 0 {
				is_obj_list = false
			} else {
				if is_obj_list {
					cols := strings.Fields(line)
					objlist.items[cols[key_index[0]]] = ClassItem{
						class_name: cols[key_index[0]],

						count: to_num(cols[key_index[1]]),
					}

				} else {
					cols := strings.Fields(line)
					objlist.total.count = to_num(cols[0])

				}
			}
		}
		line_num++
	})
	return objlist
}

type ObjListCompareResult struct {
	difference map[string]int
	added      map[string]int
	removed    []string
}

func compare_objlist(left, right ObjList) string {
	if left.class_name != right.class_name {
		return ""
	}
	var result ObjListCompareResult

	for key, right := range right.items {
		var left_value, ok = left.items[key]
		if !ok {
			result.added[key] = right.count
		} else {
			if left_value.count != right.count {
				result.difference[key] = right.count - left_value.count
			}
		}
	}

	for key := range left.items {
		if _, ok := right.items[key]; !ok {
			result.removed = append(result.removed, key)
		}
	}
	jsonData, err := json.Marshal(result)
	if err != nil {
		return ""
	}
	return string(jsonData)
}
