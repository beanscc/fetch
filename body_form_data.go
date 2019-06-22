package fetch

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
)

// FormData form-data
type FormDataBody struct {
	Param       map[string]string // 表单字段（不含文件）
	Files       []File            // 表单文件
	contentType string            // 表单 content-type 头信息
}

// File 文件
type File struct {
	Field string // 表单字段
	Name  string // 文件路径和名称
}

// NewFormDataBody return new FormDataBody
func NewFormDataBody(p map[string]string, fs ...File) *FormDataBody {
	return &FormDataBody{
		Param: p,
		Files: fs,
	}
}

// Body 构建 form-data 格式的消息体
func (fd *FormDataBody) Body() (io.Reader, error) {
	// 构造 form-data
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	defer w.Close()

	// content-type
	fd.contentType = w.FormDataContentType()

	// 表单参数
	for k, v := range fd.Param {
		if err := w.WriteField(k, v); err != nil {
			// todo log
			return nil, err
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

// ContentType 构建 form-data 格式的 header 头
func (fd *FormDataBody) ContentType() string {
	return fd.contentType
}
