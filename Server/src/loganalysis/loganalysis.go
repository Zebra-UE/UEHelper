package loganalysis

import (
	"UEHelper/src/task"
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type FLogAnalysis struct {
	FileItem []FFileItem
	NameHash map[string]int
}

var instance *FLogAnalysis
var once sync.Once

func getInstance() *FLogAnalysis {
	once.Do(func() {
		instance = &FLogAnalysis{
			FileItem: make([]FFileItem, 0),
			NameHash: make(map[string]int),
		}
	})
	return instance
}

func loadRemarks(path string, name string) string {
	remarkPath := filepath.Join(path, "remark.txt")
	if _, exists := os.Stat(remarkPath); exists != nil {
		return ""
	}
	content, err := os.ReadFile(remarkPath)
	if err != nil {
		fmt.Printf("Failed to read remark file: %v\n", err)
		return ""
	}
	return string(content)
}

type FFileItem struct {
	Index       int    `json:"index"`
	Name        string `json:"name"`
	Remark      string `json:"remark"`
	Path        string
	LogFileName string
	UnZipPath   string
	IsDecrypted bool
}

func unzip(src string) (string, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return "", err
	}
	defer r.Close()
	dest := strings.TrimSuffix(src, ".zip")
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

func View(ctx *gin.Context) {

	items := List("C:/Users/36038/Downloads")
	ctx.HTML(200, "loganalysis.html", gin.H{
		"items": items,
	})
}

func AnalyzeAPI(ctx *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	if index, ok := getInstance().NameHash[req.Name]; ok {
		fileItem := &getInstance().FileItem[index]
		if fileItem.Remark != "" {
			ctx.JSON(200, gin.H{
				"remark": fmt.Sprintf("%s", fileItem.Remark),
			})
			return
		} else {
			if fileItem.UnZipPath == "" {
				unzipPath, err := unzip(fileItem.Path)
				if err != nil {
					ctx.JSON(500, gin.H{"error": "Failed to unzip file: " + err.Error()})
					return
				}
				fileItem.UnZipPath = unzipPath
			}
			if fileItem.LogFileName == "" && fileItem.UnZipPath != "" {
				fileItem.LogFileName = findLogFile(fileItem.UnZipPath)
			}
			var logPath string
			if fileItem.IsDecrypted {
				logPath = filepath.Join(fileItem.UnZipPath, "S1Game_Decrypt.log")
			} else {
				logPath = filepath.Join(fileItem.UnZipPath, fileItem.LogFileName)
				if IsEncrypted(logPath) {
					// 如果不是S1Game.log的命名，就把它复制一份并命名为S1Game.log，因为DecryptLog.exe只能识别这个名字
					if fileItem.LogFileName != "S1Game.log" {
						standardLogPath := filepath.Join(fileItem.UnZipPath, "S1Game.log")
						if err := os.Link(logPath, standardLogPath); err != nil {
							ctx.JSON(500, gin.H{"error": "Failed to create link for log file: " + err.Error()})
							return
						}
						logPath = standardLogPath
					}
					decryptPath, err := Decrypt(fileItem.UnZipPath)
					if err != nil {
						ctx.JSON(500, gin.H{"error": "Failed to decrypt file: " + err.Error()})
						return
					}
					logPath = decryptPath
					fileItem.IsDecrypted = true
				}
			}

			remark := TryGenerateRemark(logPath)

			fileItem.Remark = remark

			ctx.JSON(200, gin.H{
				"remark": fmt.Sprintf("%s", remark),
			})
			return
		}
	}

	ctx.JSON(200, gin.H{
		"remark": fmt.Sprintf("%s", ""),
	})
}

func TryGenerateRemark(logPath string) string {
	f, err := os.Open(logPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var remark string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "An uncompress error occurred") {
			remark = "Oodle问题，解压问题"
			break
		}
		if strings.Contains(line, "OodleLZ_Decompress failed") {
			remark = "Oodle问题，解压失败"
			break
		}
	}

	if remark != "" {
		remarkPath := filepath.Join(filepath.Dir(logPath), "remark.txt")
		if err := os.WriteFile(remarkPath, []byte(remark), 0644); err != nil {
			fmt.Printf("Failed to write remark file: %v\n", err)
		}
	}
	return remark
}

func Decrypt(workPath string) (string, error) {
	// 1.  Copy DecryptLog.exe to workPath
	// Assuming DecryptLog.exe is in the same directory as the executable or a known tools path
	// For now, let's assume it's in "tools/DecryptLog.exe" relative to the server root
	exeSourcePath := "C:/Users/36038/Downloads/DecryptLog.exe" // Adjust this source path as needed
	destExePath := filepath.Join(workPath, "DecryptLog.exe")

	// Open source file
	sourceFile, err := os.Open(exeSourcePath)
	if err != nil {
		// If copy fails or file doesn't exist, we might try to run it directly if it already exists there or try another path
		// But let's stick to the requirement: copy to path
		return "", fmt.Errorf("failed to open source DecryptLog.exe: %w", err)
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(destExePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination DecryptLog.exe: %w", err)
	}
	defer destFile.Close()

	// Copy content
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy DecryptLog.exe: %w", err)
	}
	// Explicitly close dest to ensure handle is released before execution
	_ = destFile.Close()

	// 2. Execute DecryptLog.exe in workPath
	cmd := exec.Command(destExePath)
	cmd.Dir = workPath // Set working directory so it finds the log file locally

	// Capture output for debugging if needed
	// output, err := cmd.CombinedOutput()
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run DecryptLog.exe: %w", err)
	}

	// 3. Check for S1Game_Decrypt.log
	decryptLogPath := filepath.Join(workPath, "S1Game_Decrypt.log")
	if _, err := os.Stat(decryptLogPath); os.IsNotExist(err) {
		return "", fmt.Errorf("S1Game_Decrypt.log not found after execution")
	}

	// 4. Check if S1Game_Decrypt.log is encrypted
	// The requirement "检查S1Game_Decrypt.log是否是加密状态" implies we check it again.
	// If IsEncrypted returns true, it failed to decrypt properly.
	if IsEncrypted(decryptLogPath) {
		return "", fmt.Errorf("S1Game_Decrypt.log is still encrypted or invalid")
	}

	return decryptLogPath, nil
}

func IsEncrypted(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true // 如果无法打开，默认认为是加密的或者有问题
	}
	defer f.Close()

	// 读取前20行
	scanner := bufio.NewScanner(f)
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// 规则1：第一行以 "Log file open" 开头
		if lineCount == 1 && strings.HasPrefix(line, "Log file open") {
			return false
		}

		// 规则2：前20行里某一行有 "Log" 字样
		if strings.Contains(line, "Log") {
			return false
		}

		if lineCount >= 20 {
			break
		}
	}

	return true
}

func isSequencedLogFile(name string) bool {
	if !strings.HasPrefix(name, "S1Game_") || !strings.HasSuffix(name, ".log") {
		return false
	}
	trimmed := strings.TrimSuffix(strings.TrimPrefix(name, "S1Game_"), ".log")
	if len(trimmed) == 0 {
		return false
	}
	if _, err := strconv.Atoi(trimmed); err != nil {
		return false
	}
	return true
}

func findLogFile(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	var sequencedCandidate string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()

		if name == "S1Game.log" {
			return name
		}
		if sequencedCandidate == "" && isSequencedLogFile(name) {
			sequencedCandidate = name
		}
	}
	return sequencedCandidate
}

func List(dirPath string) []FFileItem {
	result := &getInstance().FileItem
	uniqueMap := &getInstance().NameHash

	selectPathTask := task.FSelectPathTask{
		Path: "C:/Users/36038/Downloads",
	}
	selectPathTask.Run()
	content, _ := json.Marshal(selectPathTask)
	listTask := task.FListFileTask{IncludeChildren: false}
	listTask.Run(string(content))
	for _, path := range listTask.Paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		fileName := info.Name()
		if idx := strings.LastIndex(fileName, "_CustomizedLogFile"); idx != -1 {

			prefix := fileName[:idx]
			if _, ok := (*uniqueMap)[prefix]; !ok {
				(*uniqueMap)[prefix] = len(*result)
				*result = append(*result, FFileItem{
					Index:       (*uniqueMap)[prefix] + 1,
					Name:        prefix,
					Remark:      "",
					IsDecrypted: false,
				})
			}
			newItem := &(*result)[(*uniqueMap)[prefix]]
			if info.IsDir() {
				newItem.UnZipPath = path
				decryptedFilePath := filepath.Join(path, "S1Game_Decrypt.log")
				if _, err := os.Stat(decryptedFilePath); err == nil {
					newItem.IsDecrypted = true
				}
				newItem.Remark = loadRemarks(path, prefix)
				if newItem.LogFileName == "" {
					newItem.LogFileName = findLogFile(path)
				}

			} else if strings.HasSuffix(info.Name(), ".zip") {
				if newItem.Path == "" {
					newItem.Path = path
				} else if len(path) < len(newItem.Path) {
					// 如果新找到的路径更短（说明可能是不带(1)的版本），则替换
					newItem.Path = path
				}
			}
		}

	}
	return *result
}
