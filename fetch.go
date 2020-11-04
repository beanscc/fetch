package fetch

import (
	"context"
	"errors"
	"fmt"
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
	req              *request                   // once req
	debug            bool                       // debug 输出请求和响应的详细信息，一般用于调试期间
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
		req:              newEmptyRequest(),
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
	nf.req = newEmptyRequest()
	nf.err = nil
	nf.ctx = context.Background()
	return nf
}

// withContext return new Fetch with ctx
func (f *Fetch) withContext(ctx context.Context) *Fetch {
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
	if len(options) == 0 {
		return f
	}

	nf := f.clone()
	for _, option := range options {
		option.Apply(nf)
	}
	return nf
}

// Get get 请求
func (f *Fetch) Get(ctx context.Context, path string) *Fetch {
	return f.Method(ctx, http.MethodGet, path)
}

// Post post 请求
func (f *Fetch) Post(ctx context.Context, path string) *Fetch {
	return f.Method(ctx, http.MethodPost, path)
}

// Put put 请求
func (f *Fetch) Put(ctx context.Context, path string) *Fetch {
	return f.Method(ctx, http.MethodPut, path)
}

// Delete del 请求
func (f *Fetch) Delete(ctx context.Context, path string) *Fetch {
	return f.Method(ctx, http.MethodDelete, path)
}

// Head 请求
func (f *Fetch) Head(ctx context.Context, path string) *Fetch {
	return f.Method(ctx, http.MethodHead, path)
}

func (f *Fetch) Method(ctx context.Context, method string, path string) *Fetch {
	if f.err != nil {
		return f
	}
	nf := f.withContext(ctx)
	nf.req.Method = method
	nf.req.URL, nf.err = util.ResolveReferenceURL(nf.baseURL, path)
	return nf
}

// Query 设置查询参数
// args 支持 key-val 对，或 map[string]interface{}，或者 key-val 对和map[string]interface{}交替组合
// Query("k1", 1, "k2", 2, map[string]interface{}{"k3": "v3"})
func (f *Fetch) Query(args ...interface{}) *Fetch {
	if f.err != nil {
		return f
	}

	if len(args) > 0 {
		if f.req.Method == "" {
			f.err = errors.New("fetch.Query: empty method, please use Get()/Post() etc. or Method() to set method")
			return f
		}
		q := f.req.URL.Query()
		for i := 0; i < len(args); {
			if m, ok := args[i].(map[string]interface{}); ok {
				for k, v := range m {
					q.Set(k, util.ToString(v))
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
				q.Set(keyStr, util.ToString(val))
			} else {
				f.err = fmt.Errorf("fetch.Query: args key-val parir key[%v] must be string type", key)
				return f
			}

			i += 2
		}
		f.req.URL.RawQuery = q.Encode()
	}
	return f
}

// AddHeader 添加 http header
// args 支持 key-val 对，或 map[string]interface{}，或者 key-val 对和map[string]interface{}交替组合
func (f *Fetch) AddHeader(args ...interface{}) *Fetch {
	if f.err != nil {
		return f
	}

	if len(args) > 0 {
		for i := 0; i < len(args); {
			if m, ok := args[i].(map[string]interface{}); ok {
				for k, v := range m {
					f.req.Header.Add(k, util.ToString(v))
				}
				i++
				continue
			}

			if i == len(args)-1 {
				f.err = errors.New("fetch.AddHeader: args must be key-val pair or map[string]interface{}")
				return f
			}

			// key-val pair
			key, val := args[i], args[i+1]
			if keyStr, ok := key.(string); ok {
				f.req.Header.Add(keyStr, util.ToString(val))
			} else {
				f.err = fmt.Errorf("fetch.AddHeader: args key-val parir key[%v] must be string type", key)
				return f
			}

			i += 2
		}
	}
	return f
}

// SetHeader 设置 http header
// args 支持 key-val 对，或 map[string]interface{}，或者 key-val 对和map[string]interface{}交替组合
func (f *Fetch) SetHeader(args ...interface{}) *Fetch {
	if f.err != nil {
		return f
	}

	if len(args) > 0 {
		for i := 0; i < len(args); {
			if m, ok := args[i].(map[string]interface{}); ok {
				for k, v := range m {
					f.req.Header.Set(k, util.ToString(v))
				}
				i++
				continue
			}

			if i == len(args)-1 {
				f.err = errors.New("fetch.SetHeader: args must be key-val pair or map[string]interface{}")
				return f
			}

			// key-val pair
			key, val := args[i], args[i+1]
			if keyStr, ok := key.(string); ok {
				f.req.Header.Set(keyStr, util.ToString(val))
			} else {
				f.err = fmt.Errorf("fetch.SetHeader: args key-val parir key[%v] must be string type", key)
				return f
			}

			i += 2
		}
	}
	return f
}

func (f *Fetch) AddCookie(cs ...*http.Cookie) *Fetch {
	for _, c := range cs {
		f.req.AddCookie(c)
	}
	return f
}

func (f *Fetch) SetBasicAuth(username, password string) *Fetch {
	f.req.SetBasicAuth(username, password)
	return f
}

func (f *Fetch) buildRequest() (*http.Request, error) {
	if f.err != nil {
		return nil, f.err
	}

	if f.req.Method == "" {
		f.err = errors.New("fetch: empty method")
		return nil, f.err
	}

	if f.req.URL.String() == "" {
		f.err = errors.New("fetch: empty url")
		return nil, f.err
	}

	// build req
	req, err := http.NewRequest(f.req.Method, f.req.URL.String(), f.req.body)
	if err != nil {
		f.err = err
		return nil, err
	}
	req = req.WithContext(f.Context())

	// clone header
	req.Header = f.cloneHeader(f.req.Header)
	return req, nil
}

func (f *Fetch) cloneHeader(h http.Header) http.Header {
	if h == nil {
		return nil
	}

	// Find total number of values.
	nv := 0
	for _, vv := range h {
		nv += len(vv)
	}
	sv := make([]string, nv) // shared backing array for headers' values
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		n := copy(sv, vv)
		h2[k] = sv[:n:n]
		sv = sv[n:]
	}
	return h2
}

// ================== set request body ==================

// Send 设置请求的 body 消息体
func (f *Fetch) Body(b body.Body) *Fetch {
	if b != nil {
		bb, err := b.Body()
		if err != nil {
			f.err = err
			return f
		}

		f.req.body = bb // set body not Body
		f.req.Header.Set(body.HeaderContentType, b.ContentType())
	}

	return f
}

// JSON 发送 application/json 格式消息
// p 支持 string/[]byte/其他类型按 json.Marshal 编码
func (f *Fetch) JSON(data interface{}) *Fetch {
	return f.Body(body.NewJSON(data))
}

// XML 发送 application/xml 格式消息
// p 支持 string/[]byte/其他类型按 xml.Marshal 编码
func (f *Fetch) XML(data interface{}) *Fetch {
	return f.Body(body.NewXML(data))
}

// Form 发送 x-www-form-urlencoded 格式消息
func (f *Fetch) Form(data map[string]interface{}) *Fetch {
	return f.Body(body.NewFormFromMap(data))
}

// MultipartForm 发送 multipart/form-data 格式消息
func (f *Fetch) MultipartForm(data map[string]interface{}, fs ...body.File) *Fetch {
	return f.Body(body.NewMultipartFormFromMap(data, fs...))
}

// ================== set request body end ==================

// ================== bind body ==================

// Bind 按已注册 bind 类型，解析 http 响应
func (f *Fetch) Bind(bind binding.Binding, v interface{}) error {
	b, ok := f.bind[bind.Name()]
	if !ok {
		return fmt.Errorf("fetch.Bind: unknown bind[%s]", bind.Name())
	}

	resp, respBody, err := f.Resp()
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("fetch.Bind: nil http.Response")
	}

	return b.Bind(resp, respBody, v)
}

// BindJSON bind http.Body with json
func (f *Fetch) BindJSON(v interface{}) error {
	return f.Bind(&binding.JSON{}, v)
}

// BindXML bind http.Body with xml
func (f *Fetch) BindXML(v interface{}) error {
	return f.Bind(&binding.XML{}, v)
}

// ================== bind body end ==================

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
	handler := func(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
		if f.debug { // debug req
			_ = dumpRequest(req, true)
		}

		resp, err := f.client.Do(req)
		if err != nil {
			return resp, nil, err
		}
		defer resp.Body.Close()

		if f.debug { // debug resp
			_ = dumpResponse(resp, true)
		}

		var b []byte
		b, resp.Body, err = util.DrainBody(resp.Body)
		return resp, b, err
	}

	resp, respBody, err := f.chainInterceptor(f.Context(), req, handler)
	return &response{resp: resp, body: respBody, err: err}
}

// Resp return http.Response, resp body, err
func (f *Fetch) Resp() (*http.Response, []byte, error) {
	res := f.do()
	return res.resp, res.body, res.err
}

// Bytes 返回http响应body消息体
func (f *Fetch) Bytes() ([]byte, error) {
	return f.do().Bytes()
}

// Text 返回http响应body消息体
func (f *Fetch) Text() (string, error) {
	return f.do().Text()
}
