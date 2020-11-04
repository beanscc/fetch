package fetch

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/beanscc/fetch/util"
)

// Handler http req handle
type Handler func(ctx context.Context, req *http.Request) (*http.Response, []byte, error)

// Interceptor 请求拦截器
// 多个 interceptor one,two,three 则执行顺序是 one,two,three 的 handler 调用前的执行流，然后是 handler, 接着是 three,two,one 中 handler 调用之后的执行流
type Interceptor func(ctx context.Context, req *http.Request, handler Handler) (*http.Response, []byte, error)

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

var (
	defaultLogInterceptorLogger = func(ctx context.Context, format string, args ...interface{}) {
		log.Printf(format, args...)
	}

	DefaultLogInterceptor = LogInterceptor(&LogInterceptorRequest{})
)

type LogInterceptorRequest struct {
	ExcludeReqHeader map[string]bool                                               // 日志不记录的请求头
	MaxReqBody       int                                                           // 日志记录请求消息体的最大字节数
	MaxRespBody      int                                                           // 日志记录响应消息体的最大字节数
	Logger           func(ctx context.Context, format string, args ...interface{}) // 日志记录的方法
}

func LogInterceptor(param *LogInterceptorRequest) Interceptor {
	return func(ctx context.Context, req *http.Request, handler Handler) (resp *http.Response, respBody []byte, err error) {
		var (
			reqBody    []byte
			logReqBody []byte
		)
		if req.Body != nil { // has body
			reqBody, req.Body, err = util.DrainBody(req.Body)
			if err != nil {
				return nil, nil, err
			}

			if param.MaxReqBody > 0 && len(reqBody) > param.MaxReqBody { // 截取 req body
				logReqBody = append(logReqBody, reqBody[:param.MaxReqBody]...)
				logReqBody = append(logReqBody, "..."...)
			} else {
				logReqBody = reqBody
			}
		}

		// copy header
		h := make(http.Header, len(req.Header)-len(param.ExcludeReqHeader))
		for k, vv := range req.Header {
			if ok := param.ExcludeReqHeader[k]; !ok {
				vv2 := make([]string, len(vv))
				copy(vv2, vv)
				h[k] = vv2
			}
		}

		start := time.Now()
		resp, respBody, err = handler(ctx, req)
		var statusCode int
		if resp != nil {
			statusCode = resp.StatusCode
		}
		end := time.Now()

		var logRespBody []byte
		if param.MaxRespBody > 0 && len(respBody) > param.MaxRespBody { // 截取 resp body
			logRespBody = append(logRespBody, respBody[:param.MaxRespBody]...)
			logRespBody = append(logRespBody, "...."...)
		} else {
			logRespBody = respBody
		}

		logger := param.Logger
		if logger == nil {
			logger = defaultLogInterceptorLogger
		}

		logger(ctx, "[Fetch] method: %s, url: %s, header: %s, body: '%s', latency: %s, status: %d, resp: '%s', err: %v",
			req.Method, req.URL.String(), h, logReqBody, end.Sub(start), statusCode, logRespBody, err)

		return resp, respBody, err
	}
}
