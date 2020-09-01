package fetch

import (
	"net/http"
	"time"

	"github.com/beanscc/fetch/binding"
)

// Option 用于设置 Fetch 属性的接口
type Option interface {
	Apply(*Fetch)
}

type optionFunc func(*Fetch)

func (o optionFunc) Apply(f *Fetch) {
	o(f)
}

// 设置 client
func Client(c *http.Client) Option {
	return optionFunc(func(f *Fetch) {
		f.client = c
	})
}

// 设置 Interceptor
func Interceptors(interceptors ...Interceptor) Option {
	return optionFunc(func(f *Fetch) {
		for _, interceptor := range interceptors {
			if interceptor == nil {
				panic("fetch: nil interceptor")
			}

			f.interceptors = append(f.interceptors, interceptor)
		}

		f.chainInterceptor = chainInterceptor(f.interceptors...)
	})
}

// Bind 设置自定义响应解析器
func Bind(binds map[string]binding.Binding) Option {
	return optionFunc(func(f *Fetch) {
		for k, v := range binds {
			f.bind[k] = v
		}
	})
}

// Timeout 设置超时时间
func Timeout(t time.Duration) Option {
	return optionFunc(func(f *Fetch) {
		f.timeout = t
	})
}

// Debug 设置 debug
func Debug(debug bool) Option {
	return optionFunc(func(f *Fetch) {
		f.debug = debug
	})
}

// Options 用于设置 Fetch 属性
type Options struct {
	Debug        bool
	Timeout      time.Duration
	Bind         map[string]binding.Binding
	Client       *http.Client
	Interceptors []Interceptor
}

func (o *Options) Apply(f *Fetch) {
	f.debug = o.Debug
	f.timeout = o.Timeout

	for k, v := range o.Bind {
		f.bind[k] = v
	}

	if o.Client != nil {
		f.client = o.Client
	}

	if len(o.Interceptors) > 0 {
		for _, interceptor := range o.Interceptors {
			if interceptor == nil {
				panic("fetch: nil interceptor")
			}

			f.interceptors = append(f.interceptors, interceptor)
		}

		f.chainInterceptor = chainInterceptor(f.interceptors...)
	}
}
