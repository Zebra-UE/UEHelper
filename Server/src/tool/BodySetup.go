package tool

import (
	"bufio"
	"os"
	"strings"
)

type BodySetupTool struct {
}

type FBodySetupSize struct {
	Path string
	Size int64
}

func readBodySetupSize(line string) FBodySetupSize {
	parts := strings.Split(value, "|")
}

func afterToken(s string) (string, bool) {
	_, after, found := strings.Cut(s, "[UBodySetup::Serialize]")
	if !found {
		return "", false
	}
	return after, true // after 就是 def（标记后的所有内容）
}

func (me BodySetupTool) run(filePath string) string {

	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if value, found := afterToken(line); found {
			sizeInfo := readBodySetupSize(value)
		}
	}

}
