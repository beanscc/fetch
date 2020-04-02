package body

import (
	"bytes"
	"io"
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
	Field   string // 表单字段
	Path    string // 文件路径名称
	Content []byte // 文件内容
}

// NewFormDataBody return new MultipartForm
func NewMultipartForm(uv url.Values, fs ...File) *MultipartForm {
	return &MultipartForm{
		param: uv,
		files: fs,
	}
}

// NewMultipartFormFromMap return new MultipartForm from map
func NewMultipartFormFromMap(m map[string]interface{}, fs ...File) *MultipartForm {
	uv := getValues(m)
	return NewMultipartForm(uv, fs...)
}

// Body 构建 multipart/form-data 格式的消息体
func (fd *MultipartForm) Body() (io.Reader, error) {
	// 构造 form-data
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	defer w.Close()

	// content-type
	fd.contentType = w.FormDataContentType()

	// 表单参数
	for k, v := range fd.param {
		for _, vv := range v {
			if err := w.WriteField(k, vv); err != nil {
				return nil, err
			}
		}
	}

	// 表单文件
	if len(fd.files) > 0 {
		for _, f := range fd.files {
			fh, err := w.CreateFormFile(f.Field, f.Path)
			if err != nil {
				return nil, err
			}

			if _, err := fh.Write(f.Content); err != nil {
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
