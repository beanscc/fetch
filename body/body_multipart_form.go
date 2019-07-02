package body

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/url"
)

// MultipartForm multipart/form-data
type MultipartForm struct {
	param       url.Values // 表单字段（不含文件）
	files       []File     // 表单文件
	contentType string     // 表单 content-type 头信息
}

// File 文件
type File struct {
	Field string // 表单字段
	Name  string // 文件路径名称
}

// NewFormDataBody return new MultipartForm
func NewMultipartForm(uv url.Values, fs ...File) *MultipartForm {
	return &MultipartForm{
		param: uv,
		files: fs,
	}
}

// NewMultipartFormFromMap return new MultipartForm from map
func NewMultipartFormFromMap(m map[string]string, fs ...File) *MultipartForm {
	uv := getValues(m)
	return NewMultipartForm(uv, fs...)
}

// Body 构建 multipart/form-data 格式的消息体
func (fd *MultipartForm) Body() (io.Reader, error) {
	// 构造 form-data
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	defer func() {
		if err := w.Close(); err != nil {
			// todo log
			// fmt.Errorf("error when closing multipart form writer: %s", err)
		}
	}()

	// content-type
	fd.contentType = w.FormDataContentType()

	// 表单参数
	for k, v := range fd.param {
		for _, vv := range v {
			if err := w.WriteField(k, vv); err != nil {
				// todo log
				return nil, err
			}
		}
	}

	// 表单文件
	if len(fd.files) > 0 {
		for _, f := range fd.files {
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

// ContentType 构建 multipart/form-data 格式的 header 头
func (fd *MultipartForm) ContentType() string {
	return fd.contentType
}
