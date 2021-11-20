package body

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

// MultipartForm multipart/form-data
type MultipartForm struct {
	data        url.Values // 表单字段（不含文件）
	files       []File     // 表单文件
	contentType string     // 表单 content-type 头信息
}

// File 文件
type File struct {
	Field       string // 表单字段
	Filename    string // 文件名称
	ContentType string // 文件 content-type；若不指定，则根据 Content 判断
	Content     []byte // 文件内容
}

// NewMultipartForm return new MultipartForm from url.Values
func NewMultipartForm(uv url.Values, fs ...File) *MultipartForm {
	return &MultipartForm{
		data:  uv,
		files: fs,
	}
}

// NewMultipartFormFromMap return new MultipartForm from map
func NewMultipartFormFromMap(m map[string]interface{}, fs ...File) *MultipartForm {
	uv := map2URLValues(m)
	return NewMultipartForm(uv, fs...)
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// CreateFormFile Create form file
func (mf *MultipartForm) CreateFormFile(w *multipart.Writer, fieldName, filename string, contentType string, fileContent []byte) (int, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(fieldName), escapeQuotes(filename)))
	if contentType == "" {
		contentType = http.DetectContentType(fileContent)
	}

	h.Set("Content-Type", contentType)
	part, err := w.CreatePart(h)
	if err != nil {
		return 0, err
	}

	return part.Write(fileContent)
}

// Body 构建 multipart/form-data 格式的消息体
func (mf *MultipartForm) Body() (io.Reader, error) {
	// 构造 form-data
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	defer w.Close()

	// content-type
	mf.contentType = w.FormDataContentType()

	// 表单参数
	for k, v := range mf.data {
		for _, vv := range v {
			if err := w.WriteField(k, vv); err != nil {
				return nil, err
			}
		}
	}

	// 表单文件
	if len(mf.files) > 0 {
		for _, f := range mf.files {
			_, err := mf.CreateFormFile(w, f.Field, f.Filename, f.ContentType, f.Content)
			if err != nil {
				return nil, err
			}
		}
	}

	return &buf, nil
}

// ContentType 构建 multipart/form-data 格式的 header 头
func (mf *MultipartForm) ContentType() string {
	return mf.contentType
}
