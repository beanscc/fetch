package body

import (
	"bytes"
	"encoding/json"
	"io"
)

// Json json body
type JSON struct {
	// param 需要json序列化的参数
	// 若类型是 string/[]byte 则，按 json 字符串处理
	// 若其他类型，则按 json 格式进行序列化
	param interface{}
}

// NewJSON return JSON
func NewJSON(p interface{}) *JSON {
	return &JSON{param: p}
}

// Body return http req body
func (j *JSON) Body() (io.Reader, error) {
	b, err := j.getJSONBytes()
	if err != nil {
		return nil, err
	}

	payload := bytes.NewReader(b)
	return payload, nil
}

func (j *JSON) getJSONBytes() ([]byte, error) {
	var b []byte
	switch j.param.(type) {
	case string:
		b = []byte(j.param.(string))
	case []byte:
		b = j.param.([]byte)
	default:
		bs, err := json.Marshal(j.param)
		if err != nil {
			return nil, err
		}
		b = bs
	}

	return b, nil
}

// ContentType return json content-type
func (j *JSON) ContentType() string {
	return MIMEJSON
}
