package binding

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
)

// Xml binding obj
type Xml struct{}

// Name name of binding obj
func (x *Xml) Name() string {
	return "xml"
}

// Bind 将 http.Response 响应解析到 out 对象中
func (x *Xml) Bind(resp *http.Response, out interface{}) error {
	if resp == nil {
		return errors.New("nil resp")
	}

	if resp.Body == nil {
		return errors.New("nil resp.Body")
	}

	if err := decodeXml(resp.Body, out); err != nil {
		return err
	}

	return nil
}

// BindBody 将响应 body 消息，解析到 out 对象
func (x *Xml) BindBody(b []byte, out interface{}) error {
	return decodeXml(bytes.NewReader(b), out)
}

func decodeXml(r io.Reader, out interface{}) error {
	decoder := xml.NewDecoder(r)
	return decoder.Decode(out)
}
