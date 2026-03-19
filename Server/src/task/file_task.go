package task

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type FSelectPathTask struct {
	Path string `json:"path"`
}

func (self *FSelectPathTask) Run() {

}

type FListFileTask struct {
	IncludeChildren bool     `json:"-"`
	EndsWith        string   `json:"-"`
	Paths           []string `json:"paths"`
}

func (self *FListFileTask) Run(input string) {
	type FInputContent struct {
		Path string `json:"path"`
	}
	var content FInputContent
	err := json.Unmarshal([]byte(input), &content)
	if err != nil {
		return
	}
	if self.IncludeChildren {

	} else {
		files, err := os.ReadDir(content.Path)
		if err != nil {
			return
		}
		for _, file := range files {
			self.Paths = append(self.Paths, filepath.Join(content.Path, file.Name()))
		}
	}
}
