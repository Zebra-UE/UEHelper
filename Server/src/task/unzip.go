package task

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FUnzipTask struct {
	Src    string
	Target string
	Sync   bool
}

func (self *FUnzipTask) Run() (string, error) {
	r, err := zip.OpenReader(self.Src)
	if err != nil {
		return "", err
	}
	defer r.Close()
	dest := self.Target
	if dest == "" {
		dest = strings.TrimSuffix(self.Src, ".zip")
	}
	for _, f := range r.File {
		// 防止 Zip Slip 漏洞 (通过 ../ 访问到目标目录之外)
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return "", fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			// 如果是目录，直接创建
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// 创建当前文件所在的目录
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return "", err
		}

		// 创建并写入目标文件
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return "", err
		}

		_, err = io.Copy(outFile, rc)

		// 确保文件被正确关闭
		outFile.Close()
		rc.Close()

		if err != nil {
			return "", err
		}
	}
	return dest, nil
}
