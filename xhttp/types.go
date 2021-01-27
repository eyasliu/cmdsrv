package xhttp

import (
	"encoding/json"
	"time"

	"github.com/eyasliu/cs"
)

type Config struct {
	MsgType       SSEMsgType // 消息类型
	HeartbeatTime time.Duration
	SIDKey        string
}

type reqMessage struct {
	sid  string
	data *cs.Request
}

type responseData struct {
	Cmd   string      `json:"cmd"`   // message command, use for route
	Seqno string      `json:"seqno"` // seq number,the request id
	Code  int         `json:"code"`  // response status code
	Msg   string      `json:"msg"`   // response status message text
	Data  interface{} `json:"data"`  // response data
}

type requestData struct {
	Cmd   string          `json:"cmd"`   // message command, use for route
	Seqno string          `json:"seqno"` // seq number,the request id
	Data  json.RawMessage `json:"data"`  // response data
}
