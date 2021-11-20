package body

import (
	"errors"
	"io"
)

// Body 构造请求的body
type Body interface {
	// Body 构造http请求body
	Body() (io.Reader, error)
	// ContentType 返回 body 体结构相应的 Header content-type 类型
	ContentType() string
}

// Body 接口实现检查
var (
	_ Body = &JSON{}
	_ Body = &XML{}
	_ Body = &Form{}
	_ Body = &MultipartForm{}
	_ Body = &errBody{}
)

type errBody struct {
	err error
}

func NewErr(e error) *errBody {
	return &errBody{err: e}
}

func (e *errBody) Body() (io.Reader, error) {
	return nil, e.Error()
}

func (e *errBody) ContentType() string {
	return ""
}

func (e *errBody) Error() error {
	if e.err == nil {
		return errors.New("wrong body")
	}
	return e.err
}
