package fetch

import (
	"net/http"
	"time"

	"github.com/beanscc/fetch/binding"
)

// Option
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

// Binds 设置绑定器
func Binds(binds map[string]binding.Binding) Option {
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
