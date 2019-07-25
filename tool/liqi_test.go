package tool

import (
	"testing"
	"bytes"
	"fmt"
	"strings"
	"io/ioutil"
)

func TestFetchLatestLiqiJson(t *testing.T) {
	if err := FetchLatestLiqiJson("../platform/majsoul/proto/lq/liqi.json"); err != nil {
		t.Fatal(err)
	}
}

func TestLiqiJsonToProto3(t *testing.T) {
	content, err := fetchLatestLiqiJson()
	if err != nil {
		t.Fatal(err)
	}
	if err := LiqiJsonToProto3(content, "../platform/majsoul/proto/lq/liqi.proto"); err != nil {
		t.Fatal(err)
	}
}

func TestGenLiqiAPI(t *testing.T) {
	content, err := fetchLatestLiqiJson()
	if err != nil {
		t.Fatal(err)
	}

	c := newConverter()
	if _, err := c.LiqiJsonToProto3(content); err != nil {
		t.Fatal(err)
	}

	protoBB := bytes.Buffer{}
	protoBB.WriteString(`// Code generated by tool/liqi_test.go. DO NOT EDIT.
package majsoul

import (
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/platform/majsoul/proto/lq"
)
`)
	for _, service := range c.rpcServiceList {
		name := service.name
		for _, method := range service.methods {
			format := `
func (c *WebSocketClient) %s(req *lq.%s) (resp *lq.%s, err error) {
	respChan := make(chan *lq.%s)
	if err = c.send(".lq.%s.%s", req, respChan); err != nil {
		return
	}
	resp = <-respChan
	if resp == nil {
		return nil, fmt.Errorf("empty response")
	}`
			if _, ok := c.messageContainError[method.responseType]; ok {
				format += `
	if resp.Error != nil {
		err = fmt.Errorf("majsoul error: %%s", resp.Error.String())
	}`
			}
			format += `
	return
}
`
			protoBB.WriteString(fmt.Sprintf(format,
				strings.Title(method.name), method.requestType, method.responseType,
				method.responseType, name, method.name))
		}
	}

	if err := ioutil.WriteFile("../platform/majsoul/liqi_api.go", protoBB.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}
