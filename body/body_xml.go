package body

import (
	"bytes"
	"encoding/xml"
	"io"
)

// XML xml body
type XML struct {
	// data 需要 xml 序列化的 body 数据
	// 若类型是 string/[]byte 则，按 xml 消息字符串处理
	// 若类型是 以上类型之外的类型，则按 xml 序列化后的字符串处理
	data interface{}
}

// NewXML return *XML
func NewXML(v interface{}) *XML {
	return &XML{v}
}

// Body return http req body
func (x *XML) Body() (io.Reader, error) {
	b, err := x.Bytes()
	if err != nil {
		return nil, err
	}

	payload := bytes.NewReader(b)
	return payload, nil
}

func (x *XML) Bytes() ([]byte, error) {
	var b []byte
	switch x.data.(type) {
	case string:
		b = []byte(x.data.(string))
	case []byte:
		b = x.data.([]byte)
	default:
		bs, err := xml.Marshal(x.data)
		if err != nil {
			return nil, err
		}
		b = bs
	}

	return b, nil
}

// ContentType return xml content-type
func (x *XML) ContentType() string {
	return MIMEXML
}
