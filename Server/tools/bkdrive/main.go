package bkdrive

import (
	"os"
	"path/filepath"
	"sync"
)

type FBkDrive struct {
	RootPath string
}

var (
	instance *FBkDrive
	once     sync.Once
)

func Get() *FBkDrive {
	once.Do(func() {
		p, _ := os.Executable()
		instance = &FBkDrive{
			RootPath: filepath.Dir(p),
		}
	})
	return instance
}

func (d *FBkDrive) GetRootPath() string {
	return d.RootPath
}
