package fetch

import (
	"context"
	"net/http"
	"time"

	"github.com/beanscc/fetch/util"
)

// Handler http req handle
type Handler func(ctx context.Context, req *http.Request) (*http.Response, []byte, error)

// Interceptor 请求拦截器
// 多个 interceptor one,two,three 则执行顺序是 one,two,three 的 handler 调用前的执行流，然后是 handler, 接着是 three,two,one 中 handler 调用之后的执行流
type Interceptor func(ctx context.Context, req *http.Request, httpHandler Handler) (*http.Response, []byte, error)

// chainInterceptor 将多个 Interceptor 合并为一个
func chainInterceptor(interceptors ...Interceptor) Interceptor {
	n := len(interceptors)
	if n > 1 {
		lastI := n - 1
		return func(ctx context.Context, req *http.Request, handler Handler) (*http.Response, []byte, error) {
			var (
				chainHandler Handler
				curI         int
			)

			chainHandler = func(currentCtx context.Context, currentReq *http.Request) (*http.Response, []byte, error) {
				if curI == lastI {
					return handler(currentCtx, currentReq)
				}
				curI++
				resp, body, err := interceptors[curI](currentCtx, currentReq, chainHandler)
				curI--
				return resp, body, err
			}

			return interceptors[0](ctx, req, chainHandler)
		}
	}

	if n == 1 {
		return interceptors[0]
	}

	// n == 0; Dummy interceptor maintained for backward compatibility to avoid returning nil.
	return func(ctx context.Context, req *http.Request, handler Handler) (*http.Response, []byte, error) {
		return handler(ctx, req)
	}
}

//func XRequestIDFromContext(ctx context.Context, key string) string {
//	if ctx == nil {
//		panic("nil ctx")
//	}
//
//	val := ctx.Value(key)
//	if v, ok := val.(string); ok {
//		return v
//	}
//
//	return GetUUID()
//}
//
//func XRequestIDInterceptor(name string) Interceptor {
//	if name == "" {
//		name = "X-Request-Id"
//	}
//	return func(ctx context.Context, req *http.Request, httpHandler Handler) (*http.Response, []byte, error) {
//		req.Header.Set(name, XRequestIDFromContext(ctx))
//		return httpHandler(ctx, req)
//	}
//}

func LogInterceptor(reqExcludeHeaderDump map[string]bool) Interceptor {
	//var reqExcludeHeaderDump = map[string]bool{
	//	"Host":              true,
	//	"Transfer-Encoding": true,
	//	"Trailer":           true,
	//	"Accept":            true,
	//	"Accept-Encoding":   true,
	//	"Connection":        true,
	//	"Cache-Control":     true,
	//	"Accept-Language":   true,
	//	"Origin":            true,
	//	"Sec-Fetch-Site":    true,
	//}

	return func(ctx context.Context, req *http.Request, httpHandler Handler) (response *http.Response, body []byte, err error) {
		var reqBody []byte
		if req.Body != nil { // has body
			rb, ob, err := util.DrainBody(req.Body)
			req.Body = ob
			if err != nil {
				return nil, rb, err
			}

			reqBody = rb
		}

		h := make(http.Header, 0)
		for k, v := range req.Header {
			if ok := reqExcludeHeaderDump[k]; !ok {
				for _, vv := range v {
					h.Set(k, vv)
				}
			}
		}
		start := time.Now()
		resp, respBody, err := httpHandler(ctx, req)
		end := time.Now()
		logger.WithContext(ctx).Infof("[Fetch-Req-Log] method: %s, url: %s, body: %s, header: %s, resp: %s, latency: %s, err: %v",
			req.Method,
			req.URL.String(),
			reqBody,
			h,
			respBody,
			end.Sub(start),
			err)
		return resp, respBody, err
	}
}
