package fetch

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

// Body body interface
type Body interface {
	// Body 返回一个 io.Reader 类型，用于构造 http.NewRequest() 方法中的 body 参数
	Body() (io.Reader, error)
	// Type 返回该种消息格式对应的请求 header 头信息
	Type() http.Header
}

// Json 根据对象构建 json 格式的消息体
type Json struct {
	// 消息对象，不支持string类型， 因为Body()中会对该数进行json编码处理
	Param interface{}
}

// Body 方法中会对 Param 参数进行一次编码
func (j Json) Body() (io.Reader, error) {
	b, err := json.Marshal(j.Param)
	if err != nil {
		return nil, err
	}

	payload := bytes.NewReader(b)
	return payload, nil
}

// Type 返回 Json 消息格式的请求 header 头信息
func (j Json) Type() http.Header {
	return jsonType()
}

func jsonType() http.Header {
	h := make(http.Header, 1)
	h.Add("Content-Type", "application/json")
	return h
}

// JsonStr 根据 json 字符串构建 body 消息体
type JsonStr struct {
	// S json 消息字符串
	S string
}

// Body 将json字符串包装成一个 io.Reader 并返回
func (j JsonStr) Body() (io.Reader, error) {
	return strings.NewReader(j.S), nil
}

// Type 返回 Json 消息格式的请求 header 头信息
func (j JsonStr) Type() http.Header {
	return jsonType()
}

// XWWWFormURLEncoded 用于构建一个 x-www-form-urlencoded 消息格式的 body 体
type XWWWFormURLEncoded struct {
	// Param 消息参数
	Param map[string]string
}

// Body 构建 x-www-form-urlencoded 消息格式的消息体
func (fd XWWWFormURLEncoded) Body() (io.Reader, error) {
	if len(fd.Param) > 0 {
		q := url.Values{}
		for key, value := range fd.Param {
			q.Add(key, value)
		}

		rawQuery := q.Encode()
		payload := strings.NewReader(rawQuery)
		return payload, nil
	}

	return nil, nil
}

// Type 构建 x-www-form-urlencoded 消息类型对应的 header 头信息
func (fd XWWWFormURLEncoded) Type() http.Header {
	h := make(http.Header, 1)
	h.Add("Content-Type", "application/x-www-form-urlencoded")
	return h
}

// FormData form-data
type FormData struct {
	Param       map[string]string // 表单字段（不含文件）
	Files       []File            // 表单文件
	contentType string            // 表单 content-type 头信息
}

// File 文件
type File struct {
	Field string // 表单字段
	Name  string // 文件路径和名称
}

// Body 构建 form-data 格式的消息体
func (fd FormData) Body() (io.Reader, error) {
	// 构造 form-data
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	defer w.Close()

	if len(fd.Param) > 0 {
		fd.contentType = w.FormDataContentType()
		// 普通表单参数
		for k, v := range fd.Param {
			if err := w.WriteField(k, v); err != nil {
				// todo log
				return nil, err
			}
		}
	}

	// 表单文件
	if len(fd.Files) > 0 {
		for _, f := range fd.Files {
			fh, err := w.CreateFormFile(f.Field, f.Name)
			if err != nil {
				// todo log
				return nil, err
			}

			fv, err := ioutil.ReadFile(f.Name)
			if err != nil {
				// todo log
				return nil, err
			}

			if _, err := fh.Write(fv); err != nil {
				// todo log
				return nil, err
			}
		}
	}

	return &buf, nil
}

// Type 构建 form-data 格式的 header 头
func (fd FormData) Type() http.Header {
	h := make(http.Header, 2)
	h.Add("Content-Type", "application/x-www-form-urlencoded")

	if fd.contentType != "" {
		h.Add("Content-Type", fd.contentType)
	}

	return h
}
