package body

import (
	"io"
	"net/url"
	"strings"
)

// Form 用于构建一个 x-www-form-urlencoded 消息格式的 body 体
type Form struct {
	// param form参数
	param url.Values
}

// NewForm return new Form
func NewForm(u url.Values) *Form {
	return &Form{param: u}
}

// NewFormFromMap return new Form from map
func NewFormFromMap(m map[string]string) *Form {
	uv := getValues(m)
	return NewForm(uv)
}

func getValues(m map[string]string) url.Values {
	uv := url.Values{}
	for k, v := range m {
		uv.Set(k, v)
	}

	return uv
}

func (f *Form) Add(key, value string) *Form {
	f.param.Add(key, value)
	return f
}

func (f *Form) Set(key, value string) *Form {
	f.param.Set(key, value)
	return f
}

// Del delete key from form
func (f *Form) Del(key string) *Form {
	f.param.Del(key)
	return f
}

// Body 构建 x-www-form-urlencoded 消息格式的消息体
func (f *Form) Body() (io.Reader, error) {
	return strings.NewReader(f.param.Encode()), nil
}

// ContentType 构建 x-www-form-urlencoded 消息类型对应的content-type头信息
func (f *Form) ContentType() string {
	return MIMEPOSTFORM
}
