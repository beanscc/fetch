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
	Body() (io.Reader, error) // return body data
	Type() http.Header        // return req body content-type
}

// Json
type Json struct {
	Param interface{}
}

func (j Json) Body() (io.Reader, error) {
	b, err := json.Marshal(j.Param)
	if err != nil {
		return nil, err
	}

	payload := bytes.NewReader(b)
	return payload, nil
}

func (j Json) Type() http.Header {
	h := make(http.Header, 1)
	h.Add("Content-Type", "application/json")
	return h
}

// JsonStr json str
type JsonStr struct {
	S string
}

func (j JsonStr) Body() (io.Reader, error) {
	return strings.NewReader(j.S), nil
}

func (j JsonStr) Type() http.Header {
	h := make(http.Header, 1)
	h.Add("Content-Type", "application/json")
	return h
}

// XWWWFormURLEncoded x-www-form-urlencoded
type XWWWFormURLEncoded struct {
	Param map[string]string
}

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

func (fd XWWWFormURLEncoded) Type() http.Header {
	h := make(http.Header, 1)
	h.Add("Content-Type", "application/x-www-form-urlencoded")
	return h
}

// FormData form-data
type FormData struct {
	Param       map[string]string // 表单字段（不含文件）
	Files       []File            // 表单文件
	contentType string            // 表单 content-type
}

// File 文件
type File struct {
	Field string // 表单字段
	Name  string // 文件路径和名称
}

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

func (fd FormData) Type() http.Header {
	h := make(http.Header, 2)
	h.Add("Content-Type", "application/x-www-form-urlencoded")

	if fd.contentType != "" {
		h.Add("Content-Type", fd.contentType)
	}

	return h
}
