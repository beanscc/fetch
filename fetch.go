package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/beanscc/fetch/binding"
	"github.com/beanscc/fetch/body"
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
	baseURL                 string
	interceptors            []Interceptor              // 拦截器
	chainInterceptorHandler InterceptorHandler         // 链式拦截器，由注册的拦截器合并而来
	onceReq                 *request                   // once req
	debug                   bool                       // debug
	err                     error                      // error
	ctx                     context.Context            // ctx
	timeout                 time.Duration              // timeout
	bind                    map[string]binding.Binding // 设置 bind 的实现对象
}

var defaultFetch = New("")

func Get(ctx context.Context, path string) *Fetch {
	return defaultFetch.Get(ctx, path)
}

func Post(ctx context.Context, path string) *Fetch {
	return defaultFetch.Post(ctx, path)
}

func Put(ctx context.Context, path string) *Fetch {
	return defaultFetch.Put(ctx, path)
}

func Delete(ctx context.Context, path string) *Fetch {
	return defaultFetch.Delete(ctx, path)
}

func Head(ctx context.Context, path string) *Fetch {
	return defaultFetch.Head(ctx, path)
}

// New return new Fetch
func New(baseURL string) *Fetch {
	return &Fetch{
		client:       http.DefaultClient,
		baseURL:      baseURL,
		interceptors: make([]Interceptor, 0),
		onceReq:      newRequest(),
		debug:        false,
		err:          nil,
		ctx:          context.Background(),
		bind: map[string]binding.Binding{
			"json": &binding.JSON{},
			"xml":  &binding.XML{},
		},
	}
}

// Clone return a clone Fetch
func (f *Fetch) Clone() *Fetch {
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
	nf := f.Clone()
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

func (f *Fetch) setInterceptor(interceptor Interceptor) {
	if strings.TrimSpace(interceptor.Name) == "" {
		panic("fetch: empty interceptor.Name")
	}

	if interceptor.Handler == nil {
		panic("fetch: nil interceptor.Handler")
	}

	for ii, vv := range f.interceptors {
		if vv.Name == interceptor.Name {
			f.interceptors[ii].Handler = interceptor.Handler // update handle
			return
		}
	}

	f.interceptors = append(f.interceptors, interceptor)
}

// SetInterceptors 注册拦截器，然后合并为一个链式拦截器
func (f *Fetch) SetInterceptors(interceptors ...Interceptor) *Fetch {
	for _, v := range interceptors {
		f.setInterceptor(v)
	}
	f.chainInterceptor()
	return f
}

// chainInterceptor 合并拦截器
func (f *Fetch) chainInterceptor() {
	interceptors := make([]InterceptorHandler, 0, len(f.interceptors))
	for _, v := range f.interceptors {
		interceptors = append(interceptors, v.Handler)
	}
	f.chainInterceptorHandler = chainInterceptor(interceptors...)
}

// SetBind 设置 bind
func (f *Fetch) SetBind(key string, bind binding.Binding) *Fetch {
	f.bind[key] = bind
	return f
}

// SetBinds 设置多个 bind
func (f *Fetch) SetBinds(b map[string]binding.Binding) *Fetch {
	for k, v := range b {
		f.SetBind(k, v)
	}

	return f
}

// Error return err
func (f *Fetch) Error() error {
	return f.err
}

// Debug 设置 Debug 模式
func (f *Fetch) Debug(debug bool) *Fetch {
	f.debug = debug
	return f
}

// Timeout set timeout
func (f *Fetch) Timeout(d time.Duration) *Fetch {
	f.timeout = d
	return f
}

// setMethod 设置 http 请求方法
func (f *Fetch) setMethod(method string) {
	f.onceReq.method = method
}

// Get get 请求
func (f *Fetch) Get(ctx context.Context, path string) *Fetch {
	nf := f.WithContext(ctx)
	nf.setMethod(http.MethodGet)
	nf.setPath(path)
	return nf
}

// Post post 请求
func (f *Fetch) Post(ctx context.Context, path string) *Fetch {
	nf := f.WithContext(ctx)
	nf.setMethod(http.MethodPost)
	nf.setPath(path)
	return nf
}

// Put put 请求
func (f *Fetch) Put(ctx context.Context, path string) *Fetch {
	nf := f.WithContext(ctx)
	nf.setMethod(http.MethodPut)
	nf.setPath(path)
	return nf
}

// Delete del 请求
func (f *Fetch) Delete(ctx context.Context, path string) *Fetch {
	nf := f.WithContext(ctx)
	nf.setMethod(http.MethodDelete)
	nf.setPath(path)
	return nf
}

// Head 请求
func (f *Fetch) Head(ctx context.Context, path string) *Fetch {
	nf := f.WithContext(ctx)
	nf.setMethod(http.MethodHead)
	nf.setPath(path)
	return nf
}

// setPath 设置 refPath
func (f *Fetch) setPath(refPath string) {
	if f.Error() != nil {
		return
	}

	f.onceReq.url, f.err = ResolveReferenceURL(f.baseURL, refPath)
}

// Query 设置单个查询参数
func (f *Fetch) Query(key, value string) *Fetch {
	f.onceReq.params[key] = value
	return f
}

// QueryMany 多个查询参数
func (f *Fetch) QueryMany(params map[string]string) *Fetch {
	for key, value := range params {
		f.onceReq.params[key] = value
	}
	return f
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
	f.Body(body.NewJSON(p))
	return f
}

// XML 发送 application/xml 格式消息
// p 支持 string/[]byte/struct/map
func (f *Fetch) XML(p interface{}) *Fetch {
	f.Body(body.NewXML(p))
	return f
}

// Form 发送 x-www-form-urlencoded 格式消息
func (f *Fetch) Form(p map[string]string) *Fetch {
	f.Body(body.NewFormFromMap(p))
	return f
}

// MultipartForm 发送 multipart/form-data 格式消息
func (f *Fetch) MultipartForm(p map[string]string, fs ...body.File) *Fetch {
	f.Body(body.NewMultipartFormFromMap(p, fs...))
	return f
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
	if f.Error() != nil {
		return newErrResp(f.Error())
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
	req, err := http.NewRequest(f.onceReq.method, f.onceReq.url.String(), bb)
	if err != nil {
		return newErrResp(err)
	}

	// handle header
	for k, v := range f.onceReq.header {
		for _, vv := range v {
			req.Header.Add(k, vv)
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
		req = req.WithContext(ctx)

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
		b, resp.Body, err = DrainBody(resp.Body)
		return resp, b, err
	}

	if f.chainInterceptorHandler == nil {
		f.chainInterceptor()
	}

	resp, b, err := f.chainInterceptorHandler(f.Context(), req, httpDoHandler)
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
	return f.Bind("json", v)
}

// BindXML bind http.Body with xml
func (f *Fetch) BindXML(v interface{}) error {
	return f.Bind("xml", v)
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
