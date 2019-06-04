package fetch

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"
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
	}
}

// RegisterInterceptor 注册拦截器
// 若 name 相同，则后面注册的 interceptor 会覆盖之前的 interceptor
func (f *Fetch) RegisterInterceptor(name string, interceptor Interceptor) {
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

// RegisterInterceptors 一次注册多个拦截器
func (f *Fetch) RegisterInterceptors(interceptors ...InterceptorHandler) {
	for _, v := range interceptors {
		f.RegisterInterceptor(v.Name, v.Interceptor)
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

// todo Clone return a clone Fetch only with client and url and err
// func (f *Fetch) Clone() *Fetch {
// 	nf := new(Fetch)
// 	// todo 考虑哪些参数需要 clone
// 	nf.client = f.client                 // client 需要公用
// 	nf.onceReq.url = f.onceReq.url       // 基础的服务地址需要公用
// 	nf.onceReq.header = f.onceReq.header // 某个服务公用的 header 需要公用
// 	nf.preDo = f.preDo                   // 服务的插件需要公用
// 	nf.afterDo = f.afterDo
// 	nf.debug = f.debug // debug 公用
// 	nf.err = nil
//
// 	return nf
// }

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

	// reset resp & resp
	nf.onceReq = newRequest()
	// nf.onceResp = nil

	// // Deep copy the baseURL
	// if f.baseURL != nil {
	// 	nfURL := new(url.URL)
	// 	*nfURL = *f.baseURL
	// 	nf.baseURL = nfURL
	// }

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

func (f *Fetch) Get(URLPath string) *Fetch {
	f.setMethod(http.MethodGet)
	f.setPath(URLPath)
	return f
}

func (f *Fetch) Post(URLPath string) *Fetch {
	f.setMethod(http.MethodPost)
	f.setPath(URLPath)
	return f
}

func (f *Fetch) Put(URLPath string) *Fetch {
	f.setMethod(http.MethodPut)
	f.setPath(URLPath)
	return f
}

func (f *Fetch) Delete(URLPath string) *Fetch {
	f.setMethod(http.MethodDelete)
	f.setPath(URLPath)
	return f
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

// QueryMap 多个查询参数
func (f *Fetch) QueryMap(params map[string]string) *Fetch {
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
func (f *Fetch) Send(body Body) *Fetch {
	if body != nil {
		f.onceReq.body = body
	}

	return f
}

// SendJson 发送json格式消息, p 不支持 json 字符串形式，
// 若需要传 json 字符串，请使用 SendJsonStr 方法
func (f *Fetch) SendJson(p interface{}) *Fetch {
	f.Send(Json{Param: p})
	return f
}

// SendJsonStr 发送 json 格式消息，以传入的 json 字符串数据为消息体
func (f *Fetch) SendJsonStr(js string) *Fetch {
	f.Send(JsonStr{js})
	return f
}

// SendForm 发送 x-www-form-urlencoded 格式消息
func (f *Fetch) SendForm(p map[string]string) *Fetch {
	f.Send(XWWWFormURLEncoded{p})
	return f
}

func (f *Fetch) handleBody() (io.Reader, error) {
	if f.onceReq.body != nil {
		b, err := f.onceReq.body.Body()
		if err != nil {
			// todo log
			return nil, err
		}

		// 替换 content-type
		for k, v := range f.onceReq.body.Type() {
			// f.onceReq.header.Del(k)
			for _, vv := range v {
				f.onceReq.header.Add(k, vv)
			}
		}

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

	if err := f.validateDo(); err != nil {
		f.err = err
		return newErrResp(f.Error())
	}

	// 处理 query 参数
	f.handleParams()

	// 处理 body
	body, err := f.handleBody()
	if err != nil {
		f.err = err
		// todo log
		return newErrResp(f.Error())
	}

	// req
	req, err := http.NewRequest(f.onceReq.method, f.onceReq.url.String(), body)
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
	// todo body 读完了，body就空了，没有内容了，考虑提供一个类似 req 读取body的 getBody() 方法
	return &response{resp: resp, body: b, err: err}
}

func debugRequest(req *http.Request, body bool) error {
	dump, err := httputil.DumpRequestOut(req, body)
	if err != nil {
		log.Printf("[Fetch-Debug] Dump request failed. err=%v", err)
		return err
	}

	log.Printf("[Fetch-Debug] %s", dump)

	return nil
}

func debugResponse(resp *http.Response, body bool) error {
	dump, err := httputil.DumpResponse(resp, body)
	if err != nil {
		log.Printf("[Fetch-Debug] Dump response failed. err=%v", err)
		return err
	}

	log.Printf("[Fetch-Debug] %s", dump)
	return nil
}
