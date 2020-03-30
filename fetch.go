package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beanscc/fetch/binding"
	"github.com/beanscc/fetch/body"
	"github.com/beanscc/fetch/util"
)

// Fetch
type Fetch struct {
	client *http.Client // client
	// client 的基础 url, 若 baseURL 带有部分 path (eg: host:port/path/), 请在 path 后跟上 "/"
	// baseURL 和 URLPath 的相对/绝对关系，请参考url.ResolveReference()
	//
	// - baseURL: host/v1/api/
	//		- 若 Get(ctx, "user/profile")，则实际请求的是 host/v1/api/user/profile
	//		- 若 Get(ctx, "/user/profile")，则实际请求的是 host/user/profile
	//		- 若 Get(ctx, "/v2/api/user/profile")，则实际请求的是 host/v2/api/user/profile
	//		- 若 Get(ctx, "../order/detail")，则实际请求的是 host/v1/order/detail
	// - baseURL: host/v1/api
	//		- 若 Get(ctx, "user/profile")，则实际请求的是 host/v1/user/profile
	//		- 若 Get(ctx, "/user/profile")，则实际请求的是 host/user/profile
	//		- 若 Get(ctx, "/v2/api/user/profile")，则实际请求的是 host/v2/api/user/profile
	//		- 若 Get(ctx, "api/user/profile")，则实际请求的是 host/v1/api/user/profile
	//		- 若 Get(ctx, "../order/detail")，则实际请求的是 host/order/detail
	baseURL          string
	interceptors     []Interceptor              // 拦截器
	chainInterceptor Interceptor                // 链式拦截器，由注册的拦截器合并而来
	onceReq          *request                   // once req
	debug            bool                       // debug
	err              error                      // error
	ctx              context.Context            // ctx
	timeout          time.Duration              // timeout
	bind             map[string]binding.Binding // 设置 bind 的实现对象
}

// New return new Fetch
func New(baseURL string, options ...Option) *Fetch {
	f := &Fetch{
		client:           http.DefaultClient,
		baseURL:          baseURL,
		interceptors:     make([]Interceptor, 0),
		chainInterceptor: chainInterceptor(),
		onceReq:          newRequest(),
		debug:            false,
		err:              nil,
		ctx:              context.Background(),
		bind: map[string]binding.Binding{
			"json": &binding.JSON{},
			"xml":  &binding.XML{},
		},
	}

	return f.WithOptions(options...)
}

// clone return a clone Fetch
func (f *Fetch) clone() *Fetch {
	nf := new(Fetch)
	*nf = *f

	// reset
	nf.onceReq = newRequest()
	nf.err = nil
	nf.ctx = context.Background()
	return nf
}

// WithContext return new Fetch with ctx
func (f *Fetch) WithContext(ctx context.Context) *Fetch {
	if ctx == nil {
		panic("fetch: nil context")
	}
	nf := f.clone()
	nf.ctx = ctx
	return nf
}

// Context return context
func (f *Fetch) Context() context.Context {
	if f.ctx == nil {
		return context.Background()
	}
	return f.ctx
}

// WithOptions 返回一个设置了新 option 的 *Fetch 对象
func (f *Fetch) WithOptions(options ...Option) *Fetch {
	nf := f.clone()
	for _, option := range options {
		option.Apply(nf)
	}

	return nf
}

// Get get 请求
func (f *Fetch) Get(ctx context.Context, refPath string) *Fetch {
	return f.Method(ctx, http.MethodGet, refPath)
}

// Post post 请求
func (f *Fetch) Post(ctx context.Context, refPath string) *Fetch {
	return f.Method(ctx, http.MethodPost, refPath)
}

// Put put 请求
func (f *Fetch) Put(ctx context.Context, refPath string) *Fetch {
	return f.Method(ctx, http.MethodPut, refPath)
}

// Delete del 请求
func (f *Fetch) Delete(ctx context.Context, refPath string) *Fetch {
	return f.Method(ctx, http.MethodDelete, refPath)
}

// Path 请求
func (f *Fetch) Patch(ctx context.Context, refPath string) *Fetch {
	return f.Method(ctx, http.MethodPatch, refPath)
}

// Options 请求
func (f *Fetch) Options(ctx context.Context, refPath string) *Fetch {
	return f.Method(ctx, http.MethodOptions, refPath)
}

// Trace 请求
func (f *Fetch) Trace(ctx context.Context, refPath string) *Fetch {
	return f.Method(ctx, http.MethodTrace, refPath)
}

// Head 请求
func (f *Fetch) Head(ctx context.Context, refPath string) *Fetch {
	return f.Method(ctx, http.MethodHead, refPath)
}

func (f *Fetch) Method(ctx context.Context, method string, refPath string) *Fetch {
	nf := f.WithContext(ctx)
	nf.setMethodPath(method, refPath)
	return nf
}

// setPath 设置 refPath
func (f *Fetch) setMethodPath(method, refPath string) {
	f.onceReq.method = method
	f.onceReq.url, f.err = util.ResolveReferenceURL(f.baseURL, refPath)
}

// Query 设置查询参数
// args 支持 key-val 对，或 map[string]interface{}，或者 key-val 对和map[string]interface{}交替组合
// Query("k1", 1, "k2", 2, map[string]interface{}{"k3": "v3"})
func (f *Fetch) Query(args ...interface{}) *Fetch {
	if f.err != nil {
		return f
	}

	for i := 0; i < len(args); {
		if m, ok := args[i].(map[string]interface{}); ok {
			for k, v := range m {
				f.onceReq.params[k] = util.ToString(v)
			}
			i++
			continue
		}

		if i == len(args)-1 {
			f.err = errors.New("fetch: query args must be key-val pair or map[string]interface{}")
			return f
		}

		// key-val pair
		key, val := args[i], args[i+1]
		if keyStr, ok := key.(string); ok {
			f.onceReq.params[keyStr] = util.ToString(val)
		} else {
			f.err = fmt.Errorf("fetch: query args key-val parir key[%v] must be string type", key)
			return f
		}

		i += 2
	}

	return f
}

// AddHeader 添加 http header
func (f *Fetch) AddHeader(key, value string) *Fetch {
	f.onceReq.header.Add(key, value)
	return f
}

// SetHeader 设置 http header
func (f *Fetch) SetHeader(key, value string) *Fetch {
	f.onceReq.header.Set(key, value)
	return f
}

// Send 设置请求的 body 消息体
func (f *Fetch) Body(body body.Body) *Fetch {
	if body != nil {
		f.onceReq.body = body
	}
	return f
}

// JSON 发送 application/json 格式消息
// p 支持 string/[]byte/struct/map
func (f *Fetch) JSON(p interface{}) *Fetch {
	return f.Body(body.NewJSON(p))
}

// XML 发送 application/xml 格式消息
// p 支持 string/[]byte/struct/map
func (f *Fetch) XML(p interface{}) *Fetch {
	return f.Body(body.NewXML(p))
}

// Form 发送 x-www-form-urlencoded 格式消息
func (f *Fetch) Form(p map[string]interface{}) *Fetch {
	return f.Body(body.NewFormFromMap(p))
}

// MultipartForm 发送 multipart/form-data 格式消息
func (f *Fetch) MultipartForm(p map[string]interface{}, fs ...body.File) *Fetch {
	return f.Body(body.NewMultipartFormFromMap(p, fs...))
}

// 处理 query 参数
func (f *Fetch) handleParams() {
	if len(f.onceReq.params) > 0 {
		q := f.onceReq.url.Query()
		for key, value := range f.onceReq.params {
			q.Add(key, value)
		}

		f.onceReq.url.RawQuery = q.Encode()
	}
}

func (f *Fetch) handleBody() (io.Reader, error) {
	if f.onceReq.body != nil {
		b, err := f.onceReq.body.Body()
		if err != nil {
			return nil, err
		}

		// 设置 content-type
		f.SetHeader(body.HeaderContentType, f.onceReq.body.ContentType())
		return b, nil
	}

	return nil, nil
}

func (f *Fetch) validateDo() error {
	if f.onceReq.method == "" {
		return errors.New("fetch: empty method")
	}

	if f.onceReq.url.String() == "" {
		return errors.New("fetch: empty url")
	}

	return nil
}

// do 执行 http 请求
func (f *Fetch) do() *response {
	if f.err != nil {
		return newErrResp(f.err)
	}

	err := f.validateDo()
	if err != nil {
		return newErrResp(err)
	}

	// 处理 query 参数
	f.handleParams()

	// 处理 body
	var bb io.Reader
	if isAllowedBody(f.onceReq.method) {
		bb, err = f.handleBody()
		if err != nil {
			return newErrResp(err)
		}
	}

	// new req
	req, err := http.NewRequestWithContext(f.Context(), f.onceReq.method, f.onceReq.url.String(), bb)
	if err != nil {
		return newErrResp(err)
	}

	// handle header
	for k, v := range f.onceReq.header {
		for _, vv := range v {
			req.Header.Set(k, vv)
		}
	}

	// timeout
	if f.timeout > 0 {
		var cancel context.CancelFunc
		f.ctx, cancel = context.WithTimeout(f.Context(), f.timeout)
		defer cancel()
	}

	// 定义 handle
	httpDoHandler := func(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
		if f.debug { // debug req
			_ = debugRequest(req, true)
		}

		resp, err := f.client.Do(req)
		if err != nil {
			return resp, nil, err
		}
		defer resp.Body.Close()

		if f.debug { // debug resp
			_ = debugResponse(resp, true)
		}

		var b []byte
		b, resp.Body, err = util.DrainBody(resp.Body)
		return resp, b, err
	}

	resp, b, err := f.chainInterceptor(f.Context(), req, httpDoHandler)
	return &response{
		resp: resp,
		body: b,
		err:  err,
	}
}

// Bind 按已注册 bind 类型，解析 http 响应
func (f *Fetch) Bind(bindType string, v interface{}) error {
	r := f.do()
	if r.err != nil {
		return r.err
	}

	if r.resp == nil {
		return errors.New("fetch: nil http.Response")
	}

	if b, ok := f.bind[bindType]; ok {
		return b.Bind(r.resp, r.body, v)
	}

	return fmt.Errorf("fetch: unknown bind type:%v", bindType)
}

// BindJSON bind http.Body with json
func (f *Fetch) BindJSON(v interface{}) error {
	return f.Bind(binding.JSON{}.Name(), v)
}

// BindXML bind http.Body with xml
func (f *Fetch) BindXML(v interface{}) error {
	return f.Bind(binding.XML{}.Name(), v)
}

// Resp return http.Response
func (f *Fetch) Resp() (*http.Response, error) {
	return f.do().Resp()
}

// Bytes 返回http响应body消息体
func (f *Fetch) Bytes() ([]byte, error) {
	return f.do().Bytes()
}

// Text 返回http响应body消息体
func (f *Fetch) Text() (string, error) {
	return f.do().Text()
}
