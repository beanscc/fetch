package fetch

import (
	"context"
	"net/http"
)

// Handler http req handle
type Handler func(ctx context.Context, req *http.Request) (*http.Response, error)

// Interceptor 请求拦截器
// 多个 interceptor one,two,three 则执行顺序是 one,two,three 的 handler 调用前的执行流，然后是 handler, 接着是 three,two,one 中 handler 调用之后的执行流
type Interceptor func(ctx context.Context, req *http.Request, handler Handler) (*http.Response, error)

// ChainInterceptor 将多个 Interceptor 合并为一个
func ChainInterceptor(interceptors ...Interceptor) Interceptor {
	n := len(interceptors)

	if n > 1 {
		lastI := n - 1
		return func(ctx context.Context, req *http.Request, handler Handler) (*http.Response, error) {
			var (
				chainHandler Handler
				curI         int
			)

			chainHandler = func(currentCtx context.Context, currentReq *http.Request) (*http.Response, error) {
				if curI == lastI {
					return handler(currentCtx, currentReq)
				}
				curI++
				resp, err := interceptors[curI](currentCtx, currentReq, chainHandler)
				curI--
				return resp, err
			}

			return interceptors[0](ctx, req, chainHandler)
		}
	}

	if n == 1 {
		return interceptors[0]
	}

	// n == 0; Dummy interceptor maintained for backward compatibility to avoid returning nil.
	return func(ctx context.Context, req *http.Request, handler Handler) (*http.Response, error) {
		return handler(ctx, req)
	}
}

// InterceptorHandler ...
type InterceptorHandler struct {
	Name        string      // name of Interceptor
	Interceptor Interceptor // Interceptor
}
