package Launch

import (
	"UEHelper/src/task"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type FLaunchManager struct {
}

type FLaunch struct {
}

func View(ctx *gin.Context) {

	ctx.HTML(200, "launch.html", nil)
}
func List(ctx *gin.Context) {
	var req struct {
		Branch string `json:"branch"`
	}
	req.Branch = ctx.Query("branch")

	var rsp struct {
		Changelist []string `json:"changelist"`
	}
	root := "E:/Game/grgame/custom"

	branchPath := root
	if req.Branch == "" {

		return
	}
	switch req.Branch {
	case "Trunk":
		branchPath = path.Join(branchPath, "trunk")
	case "Release":
		branchPath = path.Join(branchPath, "release")
	default:
		return
	}
	branchPath = path.Join(branchPath, "S1Game")
	files, err := os.ReadDir(branchPath)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to read directory"})
		return
	}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		rsp.Changelist = append(rsp.Changelist, file.Name())

	}

	ctx.JSON(200, rsp)
}

func copyFile(src, dst, name string) error {
	for _, ext := range []string{".exe", ".pdb"} {
		srcPath := path.Join(src, name+ext)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}
		toRemove := path.Join(dst, name+ext)
		if _, err := os.Stat(toRemove); os.IsExist(err) {
			os.Remove(toRemove)
		}
		if err := os.Rename(srcPath, toRemove); err == nil {
			return nil
		}
		in, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.Create(toRemove)
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err = io.Copy(out, in); err != nil {
			return err
		}
		out.Close()
		in.Close()

	}

	return nil
}

func start(target string, name string, withProfileParam bool) (int, error) {
	gamePath := path.Join(target, name+".exe")
	if _, err := os.Stat(gamePath); os.IsNotExist(err) {
		return 0, err
	}
	params := []string{"S1Game", "-ExecCmds=r.ShaderPipelineCache.StartupMode 1, networkversionoverrideCustom 0",
		"-StartFromLaunch", "--acecid=10024", "-nod3ddebug"}
	if withProfileParam {
		params = append(params, "-trace=gpu,cpu,llm,memory,loadtime,frame,log,bookmark,task,contextswitch")
		params = append(params, "-statnamedevents")
		params = append(params, "-llm")
	}

	cmd := exec.Command(path.Join(target, name+".exe"), params...)
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	pid := cmd.Process.Pid

	return pid, nil

}

// 升级 HTTP 连接为 WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 开发环境允许所有来源，生产环境需校验
	},
}

func Run(ctx *gin.Context) {

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	type req struct {
		Branch     string `json:"branch"`
		Config     string `json:"config"`
		Changelist string `json:"changelist"`
		Options    struct {
			RecordPerf bool `json:"recordPerf,omitempty"`
			EnableLog  bool `json:"enableLog,omitempty"`
			Replace    bool `json:"replace,omitempty"`
		} `json:"options,omitempty"`
	}
	type rsp struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}

	root := "E:/Game/grgame/custom"

	for {
		var request req
		if err := conn.ReadJSON(&request); err != nil {
			break
		}
		SourcePath := ""
		switch request.Branch {
		case "Trunk":
			root = path.Join(root, "trunk")
			SourcePath = "D:/Tk"
		case "Release":
			root = path.Join(root, "release")
			SourcePath = "D:/RLS"
		default:
			conn.WriteJSON(rsp{Type: "error", Payload: json.RawMessage(`"Invalid branch ` + request.Branch + `"`)})
			return
		}

		root = path.Join(root, "S1Game", request.Changelist, "Win64")

		type history struct {
		}
		historyPath := path.Join(root, "history.txt")
		if _, err := os.Stat(historyPath); os.IsNotExist(err) {
			os.WriteFile(historyPath, []byte(request.Changelist), 0644)
		}

		launcherPath := path.Join(root, "FateTrigger.exe")

		type TaskResult struct {
			msg string
			err error
		}
		ch := make(chan TaskResult, 1)
		var taskResult TaskResult

		if _, err := os.Stat(launcherPath); os.IsNotExist(err) {
			files, _ := os.ReadDir(root)
			var unzipTask task.FUnzipTask

			for _, file := range files {
				if file.IsDir() {
					continue
				}

				if path.Ext(file.Name()) == ".zip" {
					unzipTask = task.FUnzipTask{
						Src:    path.Join(root, file.Name()),
						Target: root,
						Sync:   true,
					}
					break
				}
			}
			if unzipTask.Src == "" {
				conn.WriteJSON(rsp{Type: "error", Payload: json.RawMessage(`"Launcher not found after unzipping"`)})

				return
			}

			go func(t task.FUnzipTask, c chan TaskResult) {
				msg, err := t.Run()
				c <- TaskResult{msg, err}
			}(unzipTask, ch)

			conn.WriteJSON(rsp{Type: "info", Payload: json.RawMessage(`"Unzipping..."`)})

			taskResult = <-ch
			if taskResult.err != nil {
				conn.WriteJSON(rsp{Type: "error", Payload: json.RawMessage(`"unzip failed: ` + taskResult.err.Error() + `"`)})
				return
			}
		}
		if _, err := os.Stat(launcherPath); os.IsNotExist(err) {
			conn.WriteJSON(rsp{Type: "error", Payload: json.RawMessage(`"Launcher not found after unzipping"`)})
			return
		}

		TargetPath := path.Join(root, "S1Game", "Binaries", "Win64")
		BinariesName := ""
		switch request.Config {
		case "Dev", "Development":
			BinariesName = "S1Game"
		case "Test":
			BinariesName = "S1Game-Win64-Test"
		case "Shipping":
			BinariesName = "S1Game-Win64-Shipping"
		}

		if request.Options.Replace {
			go func(c chan TaskResult) {
				copyFile(path.Join(SourcePath, "S1Game", "Binaries", "Win64"), TargetPath, BinariesName)
				c <- TaskResult{"", nil}
			}(ch)

			conn.WriteJSON(rsp{Type: "info", Payload: json.RawMessage(`"Replacing binaries..."`)})

			taskResult = <-ch
		}

		processId, err := start(TargetPath, BinariesName, request.Options.RecordPerf)
		if err != nil {
			conn.WriteJSON(rsp{Type: "error", Payload: json.RawMessage(`"Failed to start process"`)})
			return
		}
		conn.WriteJSON(rsp{Type: "success", Payload: json.RawMessage(`{"processId":` + strconv.Itoa(processId) + `}`)})
	}

}
