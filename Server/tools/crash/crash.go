package crash

import (
	"UEHelper/tools/factory"
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func unZip(src, dest string) error {
	// 1. 打开 ZIP 文件进行读取
	if path.Ext(src) == ".7z" || path.Ext(src) == ".gz" {
		//dest = filepath.Dir(dest)
		cmd := exec.Command("C:/Program Files/7-Zip/7z.exe", "x", "-o"+dest, src)
		err := cmd.Run()
		return err
	}
	if path.Ext(src) == ".zip" {
		r, err := zip.OpenReader(src)
		if err != nil {
			return err
		}
		defer r.Close()

		// 2. 确保目标目录存在
		if err := os.MkdirAll(dest, 0755); err != nil {
			return err
		}

		// 3. 遍历 ZIP 文件中的每一个文件
		for _, f := range r.File {
			rc, err := f.Open()
			if err != nil {
				return err
			}

			// 确保处理完当前文件后关闭读取器
			defer rc.Close()

			// 构建目标文件的完整路径
			fpath := filepath.Join(dest, f.Name)

			// 安全性检查：防止 ZIP 路径遍历攻击 (Zip Slip)
			// 确保解压路径是在目标目录内
			if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
				return fmt.Errorf("非法文件路径: %s", fpath)
			}

			// 4. 如果是目录，创建目录并继续下一个文件
			if f.FileInfo().IsDir() {
				os.MkdirAll(fpath, f.Mode())
				continue
			}

			// 5. 如果是文件，确保其父目录存在
			if err = os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return err
			}

			// 6. 创建目标文件
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			// 确保文件写入后关闭
			defer outFile.Close()

			// 7. 将源文件内容拷贝到目标文件中
			_, err = io.Copy(outFile, rc)

			// 注意：defer outFile.Close() 和 defer rc.Close() 可能会导致在循环内出错时出现问题。
			// 在生产代码中，更安全的做法是在每次文件操作后立即关闭。
			// 这里为了简洁演示，暂时保留 defer，但请注意其局限性。
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Run(dmpFile string) string {
	var zipfile string
	if path.Ext(dmpFile) == ".zip" {
		zipfile = dmpFile
		dmpFile = dmpFile[:len(dmpFile)-4]
	} else if path.Ext(dmpFile) == ".gz" && strings.HasSuffix(dmpFile, ".dmp.gz") {
		zipfile = dmpFile
		dmpFile = dmpFile[:len(dmpFile)-7]
	} else if path.Ext(dmpFile) == ".7z" {
		zipfile = dmpFile
		dmpFile = dmpFile[:len(dmpFile)-3]
	}
	if len(zipfile) == 0 {
		return ""
	}
	_, err := os.Stat(dmpFile)
	if err != nil {
		if os.IsNotExist(err) {
			unZip(zipfile, dmpFile)
		}
	}

	var crashFile string

	err = filepath.WalkDir(dmpFile, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".dmp") {
			crashFile = path
		}
		return nil

	})
	if err != nil {
		return ""
	}

	outputPath := filepath.Join(dmpFile, "result.log")

	_, err = os.Stat(outputPath)
	if err != nil && os.IsNotExist(err) {
		cdb := "C:/Program Files (x86)/Windows Kits/10/Debuggers/x64/cdb.exe"
		//symbolpath := fmt.Sprintf("cache*%s;", "C:/mysymbol")
		var symbolpath string
		//SRV*C:\mysymbol*http://msdl.microsoft.com/download/symbols;Y:\SymStore
		symbolpath = fmt.Sprintf("SRV*%s*http://msdl.microsoft.com/download/symbols;%s;", "C:/mysymbol", "Y:/SymStore")
		symbolpath = "C:mysymbol;Y:/SymStore;SRV*https://msdl.microsoft.com/download/symbols"
		//symbolpath = fmt.Sprintf("cache*%s;", "C:/mysymbol")
		//symbolpath = symbolpath + fmt.Sprintf("srv*%s*%s;", "C:/mysymbol", "Y:/SymStore")
		//symbolpath = symbolpath + fmt.Sprintf("srv*%s*http://msdl.microsoft.com/download/symbols;", "C:/mysymbol")
		cmd := exec.Command(cdb, "-z", crashFile, "-lines", "-c", "!analyze -v;q", "-logo", outputPath, "-y", symbolpath)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err = cmd.Run()

		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			return ""
		}
	}

	_, err = os.Stat(outputPath)
	if err != nil && os.IsNotExist(err) {
		return ""
	}

	stackText := make([]string, 0)
	beginStack := false
	factory.ReadLines(outputPath, func(s string) bool {
		if beginStack {
			if (len(s)) == 0 {
				beginStack = false
				return false
			} else {
				arr := strings.Split(s, ":")
				if len(arr) >= 2 {
					stackText = append(stackText, strings.Join(arr[2:], ":"))
				}

			}
		} else if strings.HasPrefix(s, "STACK_TEXT") {
			beginStack = true
		}
		return true
	})

	return strings.Join(stackText, "\n")

}

func List(absPath string) string {
	entries, err := os.ReadDir(absPath)
	crashHref := make([]string, 0)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".dmp.gz") || strings.HasPrefix(entry.Name(), "UECC-Windows") {
			if strings.Contains(entry.Name(), "(") || strings.Contains(entry.Name(), ")") {
				continue
			}

			crashHref = append(crashHref, fmt.Sprintf("<a href=crash/%s>%s</a>", filepath.Base(entry.Name()), entry.Name()))
		}
	}
	return strings.Join(crashHref, "\n")
}
