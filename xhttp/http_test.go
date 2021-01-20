package xhttp_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/eyasliu/cmdsrv"
	"github.com/eyasliu/cmdsrv/xhttp"
	"github.com/gogf/gf/test/gtest"
)

func sendToHttp(url string, r interface{}) (map[string]interface{}, error) {
	bt, _ := json.Marshal(r)
	client := http.DefaultClient
	resp, err := client.Post(url, "application/json", bytes.NewReader(bt))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 300 {
		return nil, errors.New("http status code fail")
	}
	resBt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res := map[string]interface{}{}
	err = json.Unmarshal(resBt, &res)
	return res, err
}

func TestHttpSrv(t *testing.T) {
	h := xhttp.New()
	http.Handle("/cmd1", h)
	http.HandleFunc("/cmd2", h.Handler)
	go http.ListenAndServe(":5679", nil)

	srv := h.Srv().Use(cmdsrv.AccessLogger("MYSRV")) // 打印请求响应日志

	gtest.C(t, func(t *gtest.T) {
		data := map[string]interface{}{
			"cmd":   "register",
			"seqno": "12345",
			"data":  "asdfgh",
		}
		srv.Handle(data["cmd"].(string), func(c *cmdsrv.Context) {
			c.OK(c.RawData)
		})

		time.Sleep(100 * time.Millisecond)

		res, err := sendToHttp("http://127.0.0.1:5679/cmd1", data)

		t.Assert(err, nil)
		t.Assert(res["cmd"], data["cmd"])
		t.Assert(res["data"], data["data"])
		t.Assert(res["seqno"], data["seqno"])
	})

}
