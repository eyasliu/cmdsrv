package xwebsocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/eyasliu/cmdsrv"

	"github.com/gorilla/websocket"
)

// WS websocket 适配器
type WS struct {
	upgrader websocket.Upgrader
	session  map[string]*Conn
	receive  chan *reqMessage
	sidCount uint64
}

var _ cmdsrv.ServerAdapter = &WS{}

// New 实例化 websocket 适配器
func New() *WS {
	return &WS{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		session: make(map[string]*Conn),
		receive: make(chan *reqMessage, 50),
	}
}

// Srv 使用该适配器创建命令消息服务
func (ws *WS) Srv() *cmdsrv.Srv {
	return cmdsrv.New(ws)
}

// Handler impl http.HandlerFunc to upgrade to websocket protocol
func (ws *WS) Handler(w http.ResponseWriter, req *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	ws.sidCount++

	sid := fmt.Sprintf("%d", ws.sidCount)

	defer ws.destroyConn(sid)
	ws.newConn(sid, conn)

	fmt.Println("connection")
}

// ServeHTTP impl http.Handler to upgrade to websocket protocol
func (ws *WS) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ws.Handler(w, req)
}

// Read 实现 cmdsrv.ServerAdapter 接口，读取消息，每次返回一条，循环读取
func (ws *WS) Read() (string, *cmdsrv.Request, error) {
	m, ok := <-ws.receive
	if !ok {
		return "", nil, errors.New("websocker server is shutdown")
	}
	return m.sid, m.data, nil
}

// Write 实现 cmdsrv.ServerAdapter 接口，给连接推送消息
func (ws *WS) Write(sid string, resp *cmdsrv.Response) error {
	conn, ok := ws.session[sid]
	if !ok {
		return errors.New("connection is already close")
	}
	return conn.Send(resp)
}

// Close 实现 cmdsrv.ServerAdapter 接口，关闭指定连接
func (ws *WS) Close(sid string) error {
	return ws.destroyConn(sid)
}

// GetAllSID 实现 cmdsrv.ServerAdapter 接口，获取当前服务所有SID，用于遍历连接
func (ws *WS) GetAllSID() []string {
	sids := make([]string, len(ws.session))
	for sid := range ws.session {
		sids = append(sids, sid)
	}
	return sids
}

// 初始化 ws 连接
func (ws *WS) newConn(sid string, conn *websocket.Conn) {
	ws.session[sid] = &Conn{
		Conn: conn,
	}
	ws.receive <- &reqMessage{msgType: websocket.TextMessage, data: &cmdsrv.Request{
		Cmd: cmdsrv.CmdConnected,
	}, sid: sid}
	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		if len(payload) == 0 { // heartbeat
			ws.receive <- &reqMessage{msgType: messageType, data: &cmdsrv.Request{
				Cmd: cmdsrv.CmdHeartbeat,
			}, sid: sid}
			continue
		}
		r := &requestData{}
		if err = json.Unmarshal(payload, r); err != nil {
			continue
		}
		ws.receive <- &reqMessage{msgType: messageType, data: &cmdsrv.Request{
			Cmd:     r.Cmd,
			Seqno:   r.Seqno,
			RawData: r.Data,
		}, sid: sid}
	}
}

// 销毁指定连接
func (ws *WS) destroyConn(sid string) error {
	conn, ok := ws.session[sid]
	if !ok {
		return errors.New("conn is already close")
	}
	err := conn.Close()
	if err != nil {
		return err
	}
	ws.receive <- &reqMessage{msgType: websocket.TextMessage, data: &cmdsrv.Request{
		Cmd: cmdsrv.CmdClosed,
	}, sid: sid}
	delete(ws.session, sid)
	return nil
}
