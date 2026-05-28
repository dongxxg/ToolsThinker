package notify_qywx

import (
	"encoding/json"
	"tools-thinker/support/util/http_util"
)

// 通知消息
// 通知对象
func SendNotify(notifyStr string, rootKey string) {
	if len(rootKey) == 0 {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	url := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=" + rootKey
	body := http_util.WithBody(GetMarkDownMsg(notifyStr))
	http_util.Post(url, body)
}

type Msg struct {
	Msgtype  string      `json:"msgtype"`
	Markdown interface{} `json:"markdown"`
}

type MarkDownMsg struct {
	Content interface{} `json:"content"`
}

func GetMarkDownMsg(msg string) []byte {
	t := &Msg{
		Msgtype:  "markdown",
		Markdown: MarkDownMsg{Content: msg},
	}
	res, _ := json.Marshal(t)
	return res
}
