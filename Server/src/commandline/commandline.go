package commandline

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 升级 HTTP 连接为 WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 开发环境允许所有来源，生产环境需校验
	},
}

func Exec() {

}

func Run(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	type InMsg struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	type OutMsg struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	for {
		var msg InMsg
		// 阻塞等待客户端发来一条消息，断线时 err != nil 退出循环
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		switch msg.Type {
		case "cmd":
			var p struct {
				Cmd  string   `json:"cmd"`
				Args []string `json:"args"`
			}
			if err := json.Unmarshal(msg.Payload, &p); err != nil {

				continue
			}
			//out, err := exec.Command(p.Cmd, p.Args...).CombinedOutput()
			if err != nil {

				continue
			}
			//conn.WriteJSON(OutMsg{Type: "output", Payload: string(out)})

		default:
			//conn.WriteJSON(OutMsg{Type: "error", Payload: "unknown type: " + msg.Type})
		}
	}
}
