package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ClientMsg struct {
	Type string `json:"type"` //setName, msg, send, roll, rename
	Data string `json:"data"`
}

type ServerMsg struct {
	Type   string `json:"type"` //ClientMsg.Type, refreshNOP
	Data   string `json:"data"`
	Status int    `json:"status"` //10000:success, 20000:failure
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var User sync.Map
var UserNum int

func main() {
	r := gin.Default()

	r.GET("ws", func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		go func(ws *websocket.Conn) {
			var name string
			for {
				mt, cmj, err := ws.ReadMessage()
				if err != nil {
					//断开连接
					if _, exist := User.Load(name); !exist {
						break
					}
					User.Delete(name)
					UserNum--
					//由于此时读取失败，mt为-1，而读取正常的情况下mt为1，所以此时下方send时mt必须使用1
					Send(true, nil, 1, &ServerMsg{
						Type:   "refreshNOP",
						Data:   strconv.Itoa(UserNum),
						Status: 10000,
					})
					Send(true, nil, 1, &ServerMsg{
						Type:   "msg",
						Data:   name + "退出了聊天室",
						Status: 10000,
					})
					break
				}

				var cm ClientMsg
				err = json.Unmarshal(cmj, &cm)
				if err != nil {
					continue
				}

				if cm.Type == "setName" {
					sm := ServerMsg{
						Type:   "setName",
						Status: 20000,
					}
					if name != "" {
						sm.Data = "警告：非法操作！已经设置过昵称"
					} else if res, err := CheckName(cm.Data); !res {
						sm.Data = err
					} else {
						UserNum++
						sm.Status = 10000
						name = cm.Data
						User.Store(name, ws)
						Send(true, nil, mt, &ServerMsg{
							Type:   "refreshNOP",
							Data:   strconv.Itoa(UserNum),
							Status: 10000,
						})
						Send(true, nil, mt, &ServerMsg{
							Type:   "msg",
							Data:   name + "进入了聊天室",
							Status: 10000,
						})
					}
					Send(false, ws, mt, &sm)
					continue
				}

				if name == "" {
					continue
				}

				sm := ServerMsg{
					Type:   cm.Type,
					Status: 20000,
				}
				switch cm.Type {
				case "send":
					sendData := cm.Data
					if sendData == "" {
						sm.Data = "发送内容不能为空"
					} else if sendData[0] == ' ' || sendData[len(sendData)-1] == ' ' || sendData[0] == '\n' || sendData[len(sendData)-1] == '\n' {
						sm.Data = "发送内容的开头和结尾不能为空格或回车"
					} else if strings.Contains(sendData, "------") {
						sm.Data = "发送内容不能含有>=6个的-"
					} else {
						sm.Status = 10000
						Send(true, nil, mt, &ServerMsg{
							Type:   "msg",
							Data:   name + "：\n" + sendData,
							Status: 10000,
						})
					}
				case "roll":
					sm.Status = 10000
					rand.Seed(time.Now().UnixNano())
					res := rand.Intn(101) //[0,101)
					Send(true, nil, mt, &ServerMsg{
						Type:   "msg",
						Data:   name + "掷出了" + strconv.Itoa(res) + "点",
						Status: 10000,
					})
				case "rename":
					newName := cm.Data
					if res, err := CheckName(newName); !res {
						sm.Data = err
					} else {
						sm.Data = newName
						sm.Status = 10000
						Send(true, nil, mt, &ServerMsg{
							Type:   "msg",
							Data:   name + "改名为" + newName,
							Status: 10000,
						})
						User.Store(newName, ws)
						User.Delete(name)
						name = newName
					}
				default:
					sm.Data = "警告：非法操作！不存在的类型"
				}
				Send(false, ws, mt, &sm)
			}
		}(ws)
	})

	_ = r.Run(":8080")
}

// Send 发送消息至客户端。toAny：true：发给所有人，false：发给ws
func Send(toAny bool, ws *websocket.Conn, mt int, sm *ServerMsg) {
	smj, _ := json.Marshal(*sm)
	if toAny {
		User.Range(func(_, ws any) bool {
			_ = ws.(*websocket.Conn).WriteMessage(mt, smj)
			return true
		})
	} else {
		_ = ws.WriteMessage(mt, smj)
	}
}

// CheckName 检查昵称是否合法。res：true：合法，false：不合法；err：不合法的原因
func CheckName(name string) (res bool, err string) {
	if l := len([]rune(name)); l < 1 || l > 16 {
		err = "昵称长度不在范围1~16内"
	} else if name[0] == ' ' || name[len(name)-1] == ' ' || name[0] == '\n' || name[len(name)-1] == '\n' {
		err = "昵称的开头和结尾不能为空格或回车"
	} else if _, exist := User.Load(name); exist {
		err = "昵称已存在"
	} else {
		res = true
	}
	return
}
