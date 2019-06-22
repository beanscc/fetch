package fetch

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// JsonBody
type JsonBody struct {
	Param interface{}
}

// NewJsonBody return JsonBody
func NewJsonBody(p interface{}) *JsonBody {
	return &JsonBody{Param: p}
}

// Body return http req body
func (j *JsonBody) Body() (io.Reader, error) {
	b, err := json.Marshal(j.Param)
	if err != nil {
		return nil, err
	}

	payload := bytes.NewReader(b)
	return payload, nil
}

func (j *JsonBody) ContentType() string {
	return MIMEJSON
}

// JsonStrBody 根据 json 字符串构建 body 消息体
type JsonStrBody struct {
	// S json 消息字符串
	S string
}

// NewJsonStrBody return JsonStrBody
func NewJsonStrBody(s string) *JsonStrBody {
	return &JsonStrBody{S: s}
}

// Body 将json字符串包装成一个 io.Reader 并返回
func (j *JsonStrBody) Body() (io.Reader, error) {
	return strings.NewReader(j.S), nil
}

// ContentType 返回 Json 消息格式的请求 header 头信息
func (j *JsonStrBody) ContentType() string {
	return MIMEJSON
}
