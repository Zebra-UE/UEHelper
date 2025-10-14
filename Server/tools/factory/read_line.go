package factory

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

func ReadLines(path string, processLine func(string)) {
	file, err := os.Open(path)

	defer file.Close()

	bom := []byte{0xEF, 0xBB, 0xBF}
	buffer := make([]byte, 3)
	_, err = file.Read(buffer)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	if !bytes.Equal(buffer, bom) {
		_, err = file.Seek(0, 0)
		if err != nil {
			fmt.Println("Error seeking file:", err)
			return
		}
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024) // 自定义缓冲区大小（64KB）
	scanner.Buffer(buf, 1024*1024)  // 设置最大行长度（1MB）
	for scanner.Scan() {
		line := scanner.Text()
		processLine(line)
	}
}
