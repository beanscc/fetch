package binding

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
)

// XML binding obj
type XML struct{}

// Name name of binding obj
func (x *XML) Name() string {
	return "xml"
}

// Bind 将 http.Response 响应解析到 out 对象中
func (x *XML) Bind(resp *http.Response, body []byte, out interface{}) error {
	if resp == nil {
		return errors.New("xml-bind:nil resp")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("xml-bind:incorrect response status code(%v)", resp.StatusCode)
	}

	if err := xml.Unmarshal(body, out); err != nil {
		return fmt.Errorf("xml-bind:%v", err)
	}

	return nil
}