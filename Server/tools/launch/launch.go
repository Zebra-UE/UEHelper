package launch

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
)

type VersionFile struct {
	Branch     string
	Changelist int
	Path       string
}

// LaunchManager 单例结构体
type LaunchManager struct {
	Versions []VersionFile
}

// 单例实例和 sync.Once
var (
	manager *LaunchManager
	once    sync.Once
)

// GetLaunchManager 获取单例实例
func GetLaunchManager() *LaunchManager {
	once.Do(func() {
		manager = &LaunchManager{
			Versions: make([]VersionFile, 0),
		}
		fmt.Println("LaunchManager 已初始化")
	})
	return manager
}

func ListVersionFile() []VersionFile {

}

var bkDrivePath = "E:\\Game\\custom\\"

func getProgramPath(Branch string, changelist int, config string) string {
	result := bkDrivePath + Branch + "\\" + "S1Game" + "\\" + strconv.Itoa(changelist) + "Win64\\Binaries\\S1Game"
	if config == "Development" {
		result += ".exe"
	} else {
		result += fmt.Sprintf("-Win64-%s.exe", config)
	}
	return result
}

func StartProcess(Branch string, changelist int, config string, args map[string]string) {
	// 程序路径
	programPath := getProgramPath(Branch, changelist, config)
	// 工作目录
	workDir := filepath.Dir(programPath)
	// 程序参数
	cmd_args := []string{} // 替换为实际参数
	for key, value := range args {
		cmd_args = append(cmd_args, fmt.Sprintf("-%s=%s", key, value))
	}
	// 创建命令对象
	cmd := exec.Command(programPath, cmd_args...)

	// 设置工作目录
	cmd.Dir = workDir

	// 启动程序
	err := cmd.Start()
	if err != nil {
		fmt.Printf("启动程序失败: %v\n", err)
		return
	}

	fmt.Printf("程序已启动，进程ID: %d\n", cmd.Process.Pid)

}
