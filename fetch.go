package fetch

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/beanscc/fetch/binding"
	"github.com/beanscc/fetch/body"
)

// Fetch
type Fetch struct {
	client       *http.Client         // client
	baseURL      string               // client 的基础 url
	interceptors []InterceptorHandler // 拦截器
	onceReq      *request             // once req
	debug        bool                 // debug
	err          error                // error
	ctx          context.Context      // ctx
	timeout      time.Duration        // timeout
	// retry // retry 可以考虑通过 interceptor 实现
	bind map[string]binding.Binding // 设置 bind 的实现对象
}

// New return new Fetch
func New(baseURL string) *Fetch {
	return &Fetch{
		client:       http.DefaultClient,
		baseURL:      baseURL,
		interceptors: make([]InterceptorHandler, 0),
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

// SetInterceptor 注册拦截器
// 若 name 相同，则后面注册的 interceptor 会覆盖之前的 interceptor
func (f *Fetch) SetInterceptor(name string, interceptor Interceptor) {
	if strings.TrimSpace(name) == "" {
		panic("interceptor's name should not be empty")
	}
	for k, v := range f.interceptors {
		if v.Name == name {
			f.interceptors[k].Interceptor = interceptor
			return
		}
	}

	f.interceptors = append(f.interceptors, InterceptorHandler{
		Name:        name,
		Interceptor: interceptor,
	})
}

// SetInterceptors 一次注册多个拦截器
func (f *Fetch) SetInterceptors(interceptors ...InterceptorHandler) {
	for _, v := range interceptors {
		f.SetInterceptor(v.Name, v.Interceptor)
	}
}

// getChainInterceptor 获取拦截器合并后的拦截器
func (f *Fetch) getChainInterceptor() Interceptor {
	interceptors := make([]Interceptor, 0, len(f.interceptors))
	for _, v := range f.interceptors {
		interceptors = append(interceptors, v.Interceptor)
	}
	// 合并拦截器
	return ChainInterceptor(interceptors...)
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

// SetBind 设置 bind
func (f *Fetch) SetBind(key string, bind binding.Binding) {
	f.bind[key] = bind
}

// SetBinds 设置多个 bind
func (f *Fetch) SetBinds(b map[string]binding.Binding) {
	for k, v := range b {
		f.SetBind(k, v)
	}
}

// Error return err
func (f *Fetch) Error() error {
	return f.err
}

// WithContext return new Fetch with ctx
func (f *Fetch) WithContext(ctx context.Context) *Fetch {
	if ctx == nil {
		panic("nil context")
	}
	nf := new(Fetch)
	*nf = *f
	nf.ctx = ctx
	nf.onceReq = newRequest()
	nf.err = nil
	return nf
}

// Context return context
func (f *Fetch) Context() context.Context {
	if f.ctx == nil {
		return context.Background()
	}

	return f.ctx
}

// Debug 开启 Debug 模式
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

// setPath 设置 URL path
func (f *Fetch) setPath(URLPath string) {
	if f.Error() != nil {
		return
	}

	u, err := url.Parse(f.baseURL)
	if u != nil {
		u.Path = path.Join(u.Path, URLPath)
	}

	f.onceReq.url = u
	f.err = err
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
	if body != nil && allowBody(f.onceReq.method) {
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
			// todo log
			return nil, err
		}

		// 设置 content-type
		f.onceReq.header.Set(body.HeaderContentType, f.onceReq.body.ContentType())

		return b, nil
	}

	return nil, nil
}

func (f *Fetch) validateDo() error {
	if f.onceReq.method == "" {
		return errors.New("empty method. please use method func first. Get()/Post() and so on")
	}

	if f.onceReq.url.String() == "" {
		return errors.New("empty url. please use method func first. Get()/Post() and so on")
	}

	return nil
}

// Do 执行 http 请求
func (f *Fetch) Do() *response {
	if f.Error() != nil {
		return newErrResp(f.Error())
	}

	err := f.validateDo()
	if err != nil {
		f.err = err
		return newErrResp(f.Error())
	}

	// 处理 query 参数
	f.handleParams()

	// 处理 body
	var bb io.Reader
	if allowBody(f.onceReq.method) {
		bb, err = f.handleBody()
		if err != nil {
			f.err = err
			// todo log
			return newErrResp(f.Error())
		}
	}

	// new req
	req, err := http.NewRequest(f.onceReq.method, f.onceReq.url.String(), bb)
	if err != nil {
		f.err = err
		// todo log
		return newErrResp(f.Error())
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
	handler := func(ctx context.Context, req *http.Request) (*http.Response, error) {
		req = req.WithContext(ctx)

		if f.debug { // debug req
			_ = debugRequest(req, true)
		}

		resp, err := f.client.Do(req)
		if err != nil {
			return resp, err
		}
		defer resp.Body.Close()

		if f.debug { // debug resp
			_ = debugResponse(resp, true)
		}

		return resp, nil
	}

	// 获取合并后的拦截器
	interceptor := f.getChainInterceptor()
	// 执行
	resp, err := interceptor(f.Context(), req, handler)
	if err != nil {
		return newErrResp(err)
	}

	var b []byte
	b, resp.Body, err = DrainBody(resp.Body)
	return &response{
		resp: resp,
		body: b,
		err:  err,
		bind: f.bind,
	}
}

// Bind bind 使用已注册的名为 bindType 的bind实现，解析 http 响应
func (f *Fetch) Bind(bindType string, v interface{}) error {
	return f.Do().Bind(bindType, v)
}

// BindJSON bind http.Body with json
func (f *Fetch) BindJSON(v interface{}) error {
	return f.Do().BindJSON(v)
}

// BindXML bind http.Body with xml
func (f *Fetch) BindXML(v interface{}) error {
	return f.Do().BindXML(v)
}

// Resp return http.Response
func (f *Fetch) Resp() (*http.Response, error) {
	return f.Do().Resp()
}
