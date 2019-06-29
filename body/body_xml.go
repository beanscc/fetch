package body

import (
	"bytes"
	"encoding/xml"
	"io"
)

// XML xml body
type XML struct {
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
