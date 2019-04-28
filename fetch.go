package fetch

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"

	"github.com/beanscc/fetch/hooks"
)

// Fetch
type Fetch struct {
	client   *http.Client
	url      *url.URL
	method   string
	header   http.Header
	body     io.Reader
	response *response
	preDo    []hooks.PreDoer
	afterDo  []hooks.AfterDoer
	debug    bool
	err      error
	ctx      context.Context
}

// New return new Fetch
func New(URL string) *Fetch {
	u, err := url.Parse(URL)
	return &Fetch{
		client:  http.DefaultClient,
		url:     u,
		header:  make(http.Header),
		preDo:   make([]hooks.PreDoer, 0),
		afterDo: make([]hooks.AfterDoer, 0),
		debug:   false,
		err:     err,
		ctx:     context.Background(),
	}
}

// Clone return a clone Fetch only with client and url and err
func (f *Fetch) Clone() *Fetch {
	nf := new(Fetch)
	nf.client = f.client
	nf.url = f.url
	nf.header = f.header
	nf.preDo = f.preDo
	nf.afterDo = f.afterDo
	nf.debug = f.debug
	nf.err = f.err

	return nf
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

	// Deep copy the URL
	if f.url != nil {
		nfURL := new(url.URL)
		*nfURL = *f.url
		nf.url = nfURL
	}

	return nf
}

// Context return context
func (f *Fetch) Context() context.Context {
	if f.ctx == nil {
		return context.Background()
	}

	return f.ctx
}

// Method 设置 http 请求方法
func (f *Fetch) Method(method string) *Fetch {
	f.setMethod(method)
	return f
}

// setMethod 设置 http 请求方法
func (f *Fetch) setMethod(method string) {
	f.method = method
}

// Debug 开启 Debug 模式
func (f *Fetch) Debug(debug bool) *Fetch {
	f.debug = debug
	return f
}

func (f *Fetch) PreDo(hooks ...hooks.PreDoer) *Fetch {
	f.preDo = append(f.preDo, hooks...)
	return f
}

func (f *Fetch) AfterDo(hooks ...hooks.AfterDoer) *Fetch {
	f.afterDo = append(f.afterDo, hooks...)
	return f
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

func (f *Fetch) Options() *Fetch {
	f.setMethod(http.MethodOptions)
	return f
}

// Path 设置 URL Path
func (f *Fetch) Path(URLPath string) *Fetch {
	f.setPath(URLPath)
	return f
}

// setPath 设置 URL path
func (f *Fetch) setPath(URLPath string) {
	if f.Error() != nil {
		return
	}

	f.url, f.err = f.url.Parse(path.Join(f.url.Path, URLPath))
}

// Query 设置单个查询参数
func (f *Fetch) Query(key, value string) *Fetch {
	if f.Error() != nil {
		return f
	}

	q := f.url.Query()
	q.Add(key, value)
	f.url.RawQuery = q.Encode()
	return f
}

// QueryMany 多个查询参数
func (f *Fetch) QueryMany(params map[string]string) *Fetch {
	if f.Error() != nil {
		return f
	}

	q := f.url.Query()
	for key, value := range params {
		q.Add(key, value)
	}

	f.url.RawQuery = q.Encode()
	return f
}

// QueryContext 设置单个查询参数，同时设置 ctx
func (f *Fetch) QueryContext(ctx context.Context, key, value string) *Fetch {
	f.Query(key, value)
	return f.WithContext(ctx)
}

// QueryManyContext 设置多个查询参数，同时设置 ctx
func (f *Fetch) QueryManyContext(ctx context.Context, params map[string]string) *Fetch {
	f.QueryMany(params)
	return f.WithContext(ctx)
}

// AddHeader 添加 http header
func (f *Fetch) AddHeader(key, value string) *Fetch {
	f.header.Add(key, value)
	return f
}

// SetHeader 设置 http header
func (f *Fetch) SetHeader(key, value string) *Fetch {
	f.header.Set(key, value)
	return f
}

// // Body 设置 http 请求 body
// func (f *Fetch) Body(body io.Reader) *Fetch {
// 	f.body = body
// 	return f
// }

// Send 发送请求
func (f *Fetch) Send(body io.Reader) *response {
	if f.Error() != nil {
		out := &response{err: f.Error()}
		f.response = out
		return out
	}

	f.body = body

	req, err := http.NewRequest(f.method, f.url.String(), f.body)
	if err != nil {
		f.err = err

		out := &response{err: err}
		f.response = out
		return out
	}

	// set header
	if len(f.header) > 0 {
		req.Header = f.header
	}

	// todo pre do
	if f.debug {
		_ = DebugRequest(req, true)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		f.err = err
		out := &response{err: err}
		f.response = out
		return out
	}
	defer resp.Body.Close()

	// todo after do

	if f.debug {
		_ = DebugResponse(resp, true)
	}

	b := resp.Body
	respBody, err := ioutil.ReadAll(b)
	out := &response{resp: resp, body: respBody, err: err}
	f.response = out
	return out
}

func DebugRequest(req *http.Request, body bool) error {
	dump, err := httputil.DumpRequestOut(req, body)
	if err != nil {
		log.Printf("[Debug-Req] Dump request failed. err=%v", err)
		return err
	}

	log.Printf("[Debug-Req] %s", dump)

	return nil
}

func DebugResponse(resp *http.Response, body bool) error {
	dump, err := httputil.DumpResponse(resp, body)
	if err != nil {
		log.Printf("[Debug-Res] Dump request failed. err=%v", err)
		return err
	}

	log.Printf("[Debug-Res] %s", dump)
	return nil
}

/*

使用：


## Get
fetch.New("http://www.xxx.com").
	Get("controller/action").
	Query("k1", "v1").
	Send().
	Error()

fetch.New("http://www.xxx.com").Get("controller/action").QueryContext(ctx, "k1", "v1").Send().Error()


fetch.Get("http://www.xxx.com").Path("controller/action").Query("k1", "v1").Send().Error()

## Post
fetch.New("http://www.xxx.com").Post().Path("controller/action").Query("k1", "v1").Body(strings.NewReader(`1=2`)).Send().Scan().Error()
*/
