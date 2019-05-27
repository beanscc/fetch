package fetch

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"time"
)

// Fetch
type Fetch struct {
	client       *http.Client    // client
	baseURL      string          // client 的基础 url
	interceptors []Interceptor   // 拦截器
	onceReq      *request        // once req
	onceResp     *response       // once resp
	debug        bool            // debug
	err          error           // error
	ctx          context.Context // ctx
	timeout      time.Duration   // timeout
	// retry // retry 可以考虑通过 interceptor 实现
}

// New return new Fetch
func New(baseURL string) *Fetch {
	return &Fetch{
		client:   http.DefaultClient,
		baseURL:  baseURL,
		onceReq:  newRequest(),
		onceResp: new(response),
		debug:    false,
		err:      nil,
		ctx:      context.Background(),
	}
}

// UseInterceptor 使用拦截器
func (f *Fetch) UseInterceptor(interceptors ...Interceptor) {
	if f.interceptors == nil {
		f.interceptors = make([]Interceptor, 0, len(interceptors))
	}

	f.interceptors = interceptors
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
	nf.onceResp = nil

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
	f.onceReq.queryParameter[key] = value
	return f
}

// QueryMap 多个查询参数
func (f *Fetch) QueryMap(params map[string]string) *Fetch {
	for key, value := range params {
		f.onceReq.queryParameter[key] = value
	}
	return f
}

// 处理 query 参数
func (f *Fetch) prepareQuery() {
	if len(f.onceReq.queryParameter) > 0 {
		q := f.onceReq.url.Query()
		for key, value := range f.onceReq.queryParameter {
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

// Send 设置 http 请求 body
func (f *Fetch) Send(body Body) *Fetch {
	if body != nil {
		f.onceReq.body = body
	}

	return f
}

// SendJSON 发送json格式, p 不支持 json 字符串形式，
// 若需要传 json 字符串，请使用 SendJSONStr 方法
func (f *Fetch) SendJSON(p interface{}) *Fetch {
	f.Send(Json{Param: p})

	return f
}

func (f *Fetch) SendJSONStr(js string) *Fetch {
	f.Send(JsonStr{js})

	return f
}

// SendForm 发送 form 格式数据
func (f *Fetch) SendForm(p map[string]string) *Fetch {
	f.Send(XWWWFormURLEncoded{p})

	return f
}

func (f *Fetch) handleBody() (io.Reader, error) {
	if f.onceReq.body != nil {
		r, err := f.onceReq.body.Body()
		if err != nil {
			// todo log
			return nil, err
		}

		// 替换 content-type
		for k, v := range f.onceReq.body.Type() {
			f.onceReq.header.Del(k)
			for _, vv := range v {
				f.onceReq.header.Add(k, vv)
			}
		}

		return r, nil
	}

	return nil, nil
}

// Do 执行 http 请求
func (f *Fetch) Do() *response {
	if f.Error() != nil {
		return newErrResp(f.Error())
	}

	// 处理 query 参数
	f.prepareQuery()

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
	if len(f.onceReq.header) > 0 {
		req.Header = f.onceReq.header
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

		// 转换 resp.Body 为 io.ReaderNoClose
		var err2 error
		resp.Body, err2 = NopCloserRespBody(resp.Body)
		return resp, err2
	}

	// 合并拦截器
	interceptor := ChainInterceptor(f.interceptors...)
	// 执行
	resp, err := interceptor(f.Context(), req, handler)
	if err != nil {
		return newErrResp(err)
	}

	var b []byte
	b, resp.Body, err = DrainBody(resp.Body)
	// todo body 读完了，body就空了，没有内容了，考虑提供一个类似 req 读取body的 getBody() 方法
	f.onceResp = &response{resp: resp, body: b, err: err}
	return f.onceResp
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
