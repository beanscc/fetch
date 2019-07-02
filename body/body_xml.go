package body

import (
	"bytes"
	"encoding/xml"
	"io"
)

// XML xml body
type XML struct {
	// param xml body 消息体
	// 若类型是 string/[]byte 则，按 xml 消息字符串处理
	// 若类型是 以上类型之外的类型，则按 xml 序列化后的字符串处理
	param interface{}
}

// NewXML return new xml with param p
func NewXML(p interface{}) *XML {
	return &XML{p}
}

// Body return http req body
func (x *XML) Body() (io.Reader, error) {
	b, err := x.getJSONBytes()
	if err != nil {
		return nil, err
	}

	payload := bytes.NewReader(b)
	return payload, nil
}

func (x *XML) getJSONBytes() ([]byte, error) {
	var b []byte
	switch x.param.(type) {
	case string:
		b = []byte(x.param.(string))
	case []byte:
		b = x.param.([]byte)
	default:
		bs, err := xml.Marshal(x.param)
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
