package body

import (
	"io"
	"net/url"
	"strings"

	"github.com/beanscc/fetch/util"
)

// Form 用于构建一个 x-www-form-urlencoded 消息格式的 body 体
type Form struct {
	// data form 数据
	data url.Values
}

// NewForm return new Form
func NewForm(u url.Values) *Form {
	return &Form{data: u}
}

// NewFormFromMap return new Form from map
func NewFormFromMap(m map[string]interface{}) *Form {
	uv := map2URLValues(m)
	return NewForm(uv)
}

func map2URLValues(m map[string]interface{}) url.Values {
	uv := url.Values{}
	for k, v := range m {
		uv.Set(k, util.ToString(v))
	}
	return uv
}

func (f *Form) Add(key string, value interface{}) *Form {
	f.data.Add(key, util.ToString(value))
	return f
}

func (f *Form) Set(key string, value interface{}) *Form {
	f.data.Set(key, util.ToString(value))
	return f
}

// Del delete key from form
func (f *Form) Del(key string) *Form {
	f.data.Del(key)
	return f
}

// Body 构建 x-www-form-urlencoded 消息格式的消息体
func (f *Form) Body() (io.Reader, error) {
	return strings.NewReader(f.Encode()), nil
}

// ContentType 构建 x-www-form-urlencoded 消息类型对应的content-type头信息
func (f *Form) ContentType() string {
	return MIMEPOSTFORM
}

// Encode return url.Values.Encode()
func (f *Form) Encode() string {
	return f.data.Encode()
}
