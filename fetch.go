package fetch

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/beanscc/fetch/binding"
	"github.com/beanscc/fetch/body"
	"github.com/beanscc/fetch/util"
)

// Fetch
type Fetch struct {
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
	client           *http.Client               // client
	interceptors     []Interceptor              // 拦截器
	chainInterceptor Interceptor                // 链式拦截器，由注册的拦截器合并而来
	onceReq          *http.Request              // once req
	debug            bool                       // debug
	err              error                      // error
	ctx              context.Context            // ctx
	timeout          time.Duration              // timeout duration
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

func newRequest() *http.Request {
	return &http.Request{
		Header: make(http.Header),
		Body:   nil,
	}
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
	f.onceReq.Method = method
	f.onceReq.URL, f.err = util.ResolveReferenceURL(f.baseURL, refPath)
}

// Query 设置查询参数
// args 支持 key-val 对，或 map[string]interface{}，或者 key-val 对和map[string]interface{}交替组合
// Query("k1", 1, "k2", 2, map[string]interface{}{"k3": "v3"})
func (f *Fetch) Query(args ...interface{}) *Fetch {
	if f.err != nil {
		return f
	}

	if len(args) > 0 {
		q := f.onceReq.URL.Query()
		for i := 0; i < len(args); {
			if m, ok := args[i].(map[string]interface{}); ok {
				for k, v := range m {
					q.Add(k, util.ToString(v))
				}
				i++
				continue
			}

			if i == len(args)-1 {
				f.err = errors.New("fetch.Query: args must be key-val pair or map[string]interface{}")
				return f
			}

			// key-val pair
			key, val := args[i], args[i+1]
			if keyStr, ok := key.(string); ok {
				q.Add(keyStr, util.ToString(val))
			} else {
				f.err = fmt.Errorf("fetch.Query: args key-val parir key[%v] must be string type", key)
				return f
			}

			i += 2
		}
		f.onceReq.URL.RawQuery = q.Encode()
	}
	return f
}

// AddHeader 添加 http header
func (f *Fetch) AddHeader(key, value string) *Fetch {
	f.onceReq.Header.Add(key, value)
	return f
}

// SetHeader 设置 http header
func (f *Fetch) SetHeader(key, value string) *Fetch {
	f.onceReq.Header.Set(key, value)
	return f
}

func (f *Fetch) AddCookie(cs ...*http.Cookie) *Fetch {
	for _, c := range cs {
		f.onceReq.AddCookie(c)
	}
	return f
}

func (f *Fetch) SetBasicAuth(username, password string) *Fetch {
	f.onceReq.SetBasicAuth(username, password)
	return f
}

func (f *Fetch) buildRequest() (*http.Request, error) {
	if f.err != nil {
		return nil, f.err
	}

	if f.onceReq.Method == "" {
		f.err = errors.New("fetch: empty method")
		return nil, f.err
	}

	if f.onceReq.URL.String() == "" {
		f.err = errors.New("fetch: empty url")
		return nil, f.err
	}

	// build req
	req, err := http.NewRequestWithContext(f.Context(), f.onceReq.Method, f.onceReq.URL.String(), f.onceReq.Body)
	if err != nil {
		return nil, err
	}

	// clone header
	req.Header = f.onceReq.Header.Clone()
	return req, err
}

// ================== set body ==================

// Send 设置请求的 body 消息体
func (f *Fetch) Body(b body.Body) *Fetch {
	if b != nil {
		bb, err := b.Body()
		if err != nil {
			f.err = err
			return f
		}

		f.onceReq.Body = ioutil.NopCloser(bb)
		f.onceReq.Header.Set(body.HeaderContentType, b.ContentType())
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

// ================== set body end ==================

// do 构造并执行 http 请求
func (f *Fetch) do() *response {
	if f.err != nil {
		return newErrResp(f.err)
	}

	// timeout must before buildRequest
	if f.timeout > 0 {
		var cancel context.CancelFunc
		f.ctx, cancel = context.WithTimeout(f.Context(), f.timeout)
		defer cancel()
	}

	req, err := f.buildRequest()
	if err != nil {
		return newErrResp(err)
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

// ================== bind body ==================

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

// ================== bind body end==================
