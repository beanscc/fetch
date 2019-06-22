package fetch

import (
	"github.com/beanscc/fetch/binding"
)

// HeaderContentType content-type header key
const HeaderContentType = "Content-Type"

// Content-Type MIME of the data formats.
const (
	MIMEJSON              = "application/json"
	MIMEXML               = "application/xml"
	MIMETEXTXML           = "text/xml"
	MIMEHTML              = "text/html"
	MIMETEXT              = "text/plain"
	MIMEPOSTFORM          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTFORM = "multipart/form-data"
)

// bindingMIME 针对不同 mime 类型的响应，指定解析方式
var bindingMIME = map[string]binding.BindingBody{
	MIMEJSON: &binding.Json{},
	MIMEXML:  &binding.Xml{},
}
