package body

import (
	"bytes"
	"encoding/json"
	"io"
)

// Json json body
type JSON struct {
	// data 需要json序列化的数据
	// 若类型是 string/[]byte 则，按 json 字符串处理
	// 若其他类型，则按 json 格式进行序列化
	data interface{}
}

// NewJSON return JSON
func NewJSON(v interface{}) *JSON {
	return &JSON{data: v}
}

// Body return http req body
func (j *JSON) Body() (io.Reader, error) {
	b, err := j.Bytes()
	if err != nil {
		return nil, err
	}

	payload := bytes.NewReader(b)
	return payload, nil
}

func (j *JSON) Bytes() ([]byte, error) {
	var b []byte
	switch j.data.(type) {
	case string:
		b = []byte(j.data.(string))
	case []byte:
		b = j.data.([]byte)
	default:
		bs, err := json.Marshal(j.data)
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
