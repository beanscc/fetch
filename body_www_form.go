package fetch

import (
	"io"
	"net/url"
	"strings"
)

// XWWWFormURLEncoded 用于构建一个 x-www-form-urlencoded 消息格式的 body 体
type XWWWFormURLEncodedBody struct {
	// Param 消息参数
	Param map[string]string
}

// NewXWWWFormURLEncodedBody return new XWWWFormURLEncodedBody
func NewXWWWFormURLEncodedBody(p map[string]string) *XWWWFormURLEncodedBody {
	return &XWWWFormURLEncodedBody{Param: p}
}

// Body 构建 x-www-form-urlencoded 消息格式的消息体
func (f *XWWWFormURLEncodedBody) Body() (io.Reader, error) {
	if len(f.Param) > 0 {
		q := url.Values{}
		for key, value := range f.Param {
			q.Add(key, value)
		}

		rawQuery := q.Encode()
		payload := strings.NewReader(rawQuery)
		return payload, nil
	}

	return nil, nil
}

// ContentType 构建 x-www-form-urlencoded 消息类型对应的content-type头信息
func (f *XWWWFormURLEncodedBody) ContentType() string {
	return MIMEPOSTFORM
}
