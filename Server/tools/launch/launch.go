package launch

import "os"

type VersionFile struct {
	Branch     string
	Changelist int
	Path       string
}

func ListVersionFile() []VersionFile {
	var result []VersionFile
	entries, err := os.ReadDir(dir)
	return result
}
