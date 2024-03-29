package fetch_test

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/beanscc/fetch"
	"github.com/beanscc/fetch/body"
	"github.com/beanscc/fetch/util"
)

type testBaseResp struct {
	Data interface{} `json:"data,omitempty" xml:"data,omitempty"`
	Code int         `json:"code" xml:"code"`
	Msg  string      `json:"msg" xml:"msg"`
}

func newTestBaseResp(data interface{}) *testBaseResp {
	return &testBaseResp{
		Data: data,
		Code: 0,
		Msg:  "ok",
	}
}

func (r testBaseResp) json() string {
	bs, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return string(bs)
}

func (r testBaseResp) xml() string {
	bs, err := xml.Marshal(r)
	if err != nil {
		panic(err)
	}
	return string(bs)
}

func TestPost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", body.MIMEJSON)
		w.Header().Add("x-request-id", fmt.Sprintf("trace-id-%d", time.Now().UnixNano()))
		out := newTestBaseResp(nil)
		fmt.Fprint(w, out.json())
	}))

	var res testBaseResp

	ctx := context.Background()
	err := fetch.Post(ctx, ts.URL+"/api/user").
		JSON(map[string]interface{}{
			"name": "ming.liu",
			"age":  18,
		}).BindJSON(&res)
	if err != nil {
		t.Errorf("TestPost failed. err:%v", err)
		return
	}
	t.Logf("TestPost res:%+v", res)
}

// go test -v -run TestFetchGet
func TestFetchGet(t *testing.T) {
	type Resp struct {
		Name   string `json:"name"`
		Age    uint8  `json:"age"`
		Addr   string `json:"address"`
		Mobile string `json:"mobile"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out := testBaseResp{
			Code: 0,
			Msg:  "ok",
			Data: &Resp{
				Name:   "ming.liu",
				Age:    20,
				Addr:   "beijing wangfujing street",
				Mobile: "+86-13800000000",
			},
		}

		res, _ := json.Marshal(out)
		w.Header().Set("content-type", body.MIMEJSON)
		w.Header().Add("x-request-id", r.Header.Get("x-request-id"))
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}))

	f := fetch.New(ts.URL,
		fetch.Debug(true),
		fetch.Interceptors(
			// fetch.LogInterceptor 会输出以下日志内容
			// 2020/06/30 16:12:22 [Fetch] method: GET, url: http://127.0.0.1:58785/api/user?id=10&name=ming, header: map[X-Request-Id:[trace-id-1593504742037996000]], body: , latency: 1.088405ms, status: 200, resp: {"data":{"name":"ming.liu","age":20,"address":"beijing wangfujing street","mobile":"+86-13800000000"},"code":0,"msg":"ok"}, err: <nil>, extra k1:v1
			fetch.LogInterceptor(&fetch.LogInterceptorRequest{
				ExcludeReqHeader: nil,
				MaxReqBody:       0,
				MaxRespBody:      0,
				Logger: func(ctx context.Context, format string, args ...interface{}) {
					v1, _ := ctx.Value("k1").(string)
					log.Printf(format+", extra k1:%v", append(args, v1)...)
				},
			}),
		))

	var data Resp
	tRes := newTestBaseResp(&data)
	ctx := context.WithValue(context.Background(), "k1", "v1")
	err := f.
		Get(ctx, "/api/user/:id/address/:id", 10, 20).
		Query("id", 10, map[string]interface{}{"name": "ming"}).
		AddHeader("x-request-id", fmt.Sprintf("trace-id-%d", time.Now().UnixNano())).
		// Bind("json", &tRes)
		// 或
		BindJSON(&tRes)
	if err != nil {
		t.Errorf("Test_Fetch_Get failed. err:%v", err)
		return
	}
	t.Logf("resp.data=%#v", data)

	// output:
	/*
		2020/06/30 16:12:22 [Fetch] GET /api/user?id=10&name=ming HTTP/1.1
		Host: 127.0.0.1:58785
		User-Agent: Go-http-client/1.1
		X-Request-Id: trace-id-1593504742037996000
		Accept-Encoding: gzip

		2020/06/30 16:12:22 [Fetch] HTTP/1.1 200 OK
		Content-Length: 122
		Content-Type: application/json
		Date: Tue, 30 Jun 2020 08:12:22 GMT
		X-Request-Id: trace-id-1593504742037996000

		{"data":{"name":"ming.liu","age":20,"address":"beijing wangfujing street","mobile":"+86-13800000000"},"code":0,"msg":"ok"}
		2020/06/30 16:12:22 [Fetch] method: GET, url: http://127.0.0.1:58785/api/user?id=10&name=ming, header: map[X-Request-Id:[trace-id-1593504742037996000]], body: '', latency: 1.088405ms, status: 200, resp: '{"data":{"name":"ming.liu","age":20,"address":"beijing wangfujing street","mobile":"+86-13800000000"},"code":0,"msg":"ok"}', err: <nil>, extra k1:v1
		--- PASS: TestFetchGet (0.00s)
		    fetch_test.go:147: resp.data=fetch_test.Resp{Name:"ming.liu", Age:0x14, Addr:"beijing wangfujing street", Mobile:"+86-13800000000"}
	*/
}

func TestFetchPostJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", body.MIMEJSON)
		w.Header().Add("x-request-id", fmt.Sprintf("trace-id-%d", time.Now().UnixNano()))
		out := newTestBaseResp(nil)
		fmt.Fprint(w, out.json())
	}))

	var res testBaseResp
	f := fetch.New(ts.URL, fetch.Debug(true), fetch.Interceptors(
		// fetch.LogInterceptor 会输出以下日志内容
		fetch.DefaultLogInterceptor,
	))

	ctx := context.Background()
	err := f.Post(ctx, "api/user").
		JSON(map[string]interface{}{
			"name": "ming.liu",
			"age":  18,
		}).BindJSON(&res)
	if err != nil {
		t.Errorf("TestFetchPostJSON failed. err:%v", err)
		return
	}
	t.Logf("TestFetchPostJSON res:%+v", res)

	// output:
	/*
		2020/06/30 16:09:59 [Fetch] POST /api/user HTTP/1.1
		Host: 127.0.0.1:58717
		User-Agent: Go-http-client/1.1
		Transfer-Encoding: chunked
		Content-Type: application/json
		Accept-Encoding: gzip

		1c
		{"age":18,"name":"ming.liu"}
		0

		2020/06/30 16:09:59 [Fetch] HTTP/1.1 200 OK
		Content-Length: 22
		Content-Type: application/json
		Date: Tue, 30 Jun 2020 08:09:59 GMT
		X-Request-Id: trace-id-1593504599030600000

		{"code":0,"msg":"ok"}
		2020/06/30 16:09:59 [Fetch] method: POST, url: http://127.0.0.1:60661/api/user, header: map[Content-Type:[application/json]], body: '{"age":18,"name":"ming.liu"}', latency: 995.441µs, status: 200, resp: '{"code":0,"msg":"ok"}', err: <nil>
		--- PASS: TestFetchPostJSON (0.00s)
		    fetch_test.go:283: TestFetchPostJSON res:{Data:<nil> Code:0 Msg:ok}
	*/
}

func TestFetchPostXML(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", body.MIMEXML)
		w.Header().Add("x-request-id", fmt.Sprintf("trace-id-%d", time.Now().UnixNano()))
		out := newTestBaseResp(nil)
		fmt.Fprint(w, out.xml())
	}))

	type User struct {
		XMLName xml.Name `xml:"user"`
		ID      string   `xml:"id,attr"`
		Name    string   `xml:"name"`
		Age     int      `xml:"age"`
		Height  float32  `xml:"height"`
	}

	ctx := context.Background()
	var res testBaseResp
	f := fetch.New(ts.URL, fetch.Debug(true))
	err := f.Post(ctx, "api/user").
		XML(&User{
			ID:     "6135200011057538",
			Name:   "si.li",
			Age:    20,
			Height: 175,
		}).BindXML(&res)
	if err != nil {
		t.Errorf("TestFetchPostXML failed. err:%v", err)
		return
	}
	t.Logf("TestFetchPostXML res:%+v", res)

	// output:
	/*
		2020/06/30 16:09:05 [Fetch] POST /api/user HTTP/1.1
		Host: 127.0.0.1:58708
		User-Agent: Go-http-client/1.1
		Transfer-Encoding: chunked
		Content-Type: application/xml
		Accept-Encoding: gzip

		56
		<user id="6135200011057538"><name>si.li</name><age>20</age><height>175</height></user>
		0

		2020/06/30 16:09:05 [Fetch] HTTP/1.1 200 OK
		Content-Length: 57
		Content-Type: application/xml
		Date: Tue, 30 Jun 2020 08:09:05 GMT
		X-Request-Id: trace-id-1593504545433384000

		<testBaseResp><code>0</code><msg>ok</msg></testBaseResp>
		--- PASS: TestFetchPostXML (0.00s)
		    fetch_test.go:340: TestFetchPostXML res:{Data:<nil> Code:0 Msg:ok}
	*/
}

func TestFetchPostForm(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", body.MIMEJSON)
		w.Header().Add("x-request-id", fmt.Sprintf("trace-id-%d", time.Now().UnixNano()))
		out := newTestBaseResp(nil)
		fmt.Fprint(w, out.json())
	}))

	ctx := context.Background()
	f := fetch.New(ts.URL, fetch.Debug(true))

	formByMap := map[string]interface{}{
		"name": "wang.wu",
		"age":  25,
	}

	resBody, err := f.Post(ctx, "api/user").
		Form(formByMap).
		Text()
	if err != nil {
		t.Errorf("TestFetchPostForm map failed. err:%v", err)
		return
	}
	t.Logf("TestFetchPostForm map resp body:%s", resBody)

	// output:
	/*
		2021/11/20 11:06:29 [Fetch] POST /api/user HTTP/1.1
		Host: 127.0.0.1:63533
		User-Agent: Go-http-client/1.1
		Content-Length: 19
		Content-Type: application/x-www-form-urlencoded
		Accept-Encoding: gzip

		age=25&name=wang.wu
		2021/11/20 11:06:29 [Fetch] HTTP/1.1 200 OK
		Content-Length: 21
		Content-Type: application/json
		Date: Sat, 20 Nov 2021 03:06:29 GMT
		X-Request-Id: trace-id-1637377589793352000

		{"code":0,"msg":"ok"}
			fetch_test.go:287: TestFetchPostForm map resp body:{"code":0,"msg":"ok"}
	*/

	formByValues := url.Values{}
	formByValues.Add("name", "xiao.ming")
	formByValues.Add("age", "20")
	resBody, err = f.Post(ctx, "api/user").
		Form(formByValues).
		Text()
	if err != nil {
		t.Errorf("TestFetchPostForm url.Values failed. err:%v", err)
		return
	}
	t.Logf("TestFetchPostForm url.Values resp body:%s", resBody)

	// output:
	/*
		2021/11/20 11:06:29 [Fetch] POST /api/user HTTP/1.1
		Host: 127.0.0.1:63533
		User-Agent: Go-http-client/1.1
		Content-Length: 21
		Content-Type: application/x-www-form-urlencoded
		Accept-Encoding: gzip

		age=20&name=xiao.ming
		2021/11/20 11:06:29 [Fetch] HTTP/1.1 200 OK
		Content-Length: 21
		Content-Type: application/json
		Date: Sat, 20 Nov 2021 03:06:29 GMT
		X-Request-Id: trace-id-1637377589793933000

		{"code":0,"msg":"ok"}
		    fetch_test.go:299: TestFetchPostForm url.Values resp body:{"code":0,"msg":"ok"}
	*/

}

func TestFetchPostMultipartForm(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", body.MIMEJSON)
		w.Header().Add("x-request-id", fmt.Sprintf("trace-id-%d", time.Now().UnixNano()))
		out := newTestBaseResp(nil)
		fmt.Fprint(w, out.json())
	}))

	ctx := context.Background()
	f := fetch.New(ts.URL, fetch.Debug(true))

	formData := map[string]interface{}{
		"name": "wang.wu",
		"age":  25,
	}

	file1 := "testdata/f1.txt"
	file1Content, err := ioutil.ReadFile(file1)
	if err != nil {
		t.Fatalf("readFile 1 failed. err=%v", err)
	}

	file2 := "testdata/f2.txt"
	file2Content, err := ioutil.ReadFile(file2)
	if err != nil {
		t.Fatalf("readFile 2 failed. err=%v", err)
	}

	formFile := []body.File{
		{
			Field:    "file-1",
			Filename: file1, // note: 若未指定文件的 content-type，则表单发送时，根据文件内容识别此文件类型，此文件的 Content-Type: text/plain; charset=utf-8
			Content:  file1Content,
		},
		{
			Field:       "file-2",
			Filename:    file2,
			ContentType: "application/octet-stream", // note: 若指定文件的 content-type，则表单发送时，此文件的Content-Type: application/octet-stream
			Content:     file2Content,
		},
	}
	resBody, err := f.Post(ctx, "api/user").
		MultipartForm(formData, formFile...).Bytes()
	if err != nil {
		t.Errorf("TestFetchPostMultipartForm failed. err:%v", err)
		return
	}
	t.Logf("TestFetchPostMultipartForm resp body:%s", resBody)

	// output:
	/*
		2020/06/30 16:18:38 [Fetch] POST /api/user HTTP/1.1
		Host: 127.0.0.1:58880
		User-Agent: Go-http-client/1.1
		Transfer-Encoding: chunked
		Content-Type: multipart/form-data; boundary=3a27c156fa0406ed5b547dc7024c0fda21a5aa40536408dd40f95c5d0552
		Accept-Encoding: gzip

		2f5
		--3a27c156fa0406ed5b547dc7024c0fda21a5aa40536408dd40f95c5d0552
		Content-Disposition: form-data; name="name"

		wang.wu
		--3a27c156fa0406ed5b547dc7024c0fda21a5aa40536408dd40f95c5d0552
		Content-Disposition: form-data; name="age"

		25
		--3a27c156fa0406ed5b547dc7024c0fda21a5aa40536408dd40f95c5d0552
		Content-Disposition: form-data; name="file-1"; filename="testdata/f1.txt"
		Content-Type: text/plain; charset=utf-8

		this is test file.
		this is test file line 2;
		--3a27c156fa0406ed5b547dc7024c0fda21a5aa40536408dd40f95c5d0552
		Content-Disposition: form-data; name="file-2"; filename="testdata/f2.txt"
		Content-Type: application/octet-stream

		this is test file2.

		this is test file line 3;
		--3a27c156fa0406ed5b547dc7024c0fda21a5aa40536408dd40f95c5d0552--

		0

		2020/06/30 16:18:38 [Fetch] HTTP/1.1 200 OK
		Content-Length: 22
		Content-Type: application/json
		Date: Tue, 30 Jun 2020 08:18:38 GMT
		X-Request-Id: trace-id-1593505118084005000

		{"code":0,"msg":"ok"}
		--- PASS: TestFetchPostMultipartForm (0.00s)
		    fetch_test.go:365: TestFetchPostMultipartForm resp body:{"code":0,"msg":"ok"}
	*/
}

func TestFetch_WithOptions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	f := fetch.New(ts.URL, fetch.Debug(true)) // f 开启 debug
	// 或
	// f = fetch.New(ts.URL, &fetch.Options{
	// 	Debug: true,
	// })

	// Get() 方法会 clone 生成一个新的 *Fetch 对象（会clone f 的一些全局属性，清空一次性请求的属性），所以 Get() 返回的 *Fetch 对象，debug 还是开启的，会输出请求日志
	f.Get(context.Background(), "path").Bytes()

	// WithOptions() 方法会 clone 生成一个新的 *Fetch 对象，然后根据 options 参数设置关闭 debug 模式， 所以此次请求不会输出请求日志
	f.WithOptions(fetch.Debug(false)).Get(context.Background(), "path2").Bytes()

	// f 还是上面 New() 生成的 *Fetch 对象
	f.Get(context.Background(), "path3").Bytes()

	// output:
	/*
		2020/06/30 10:31:59 [Fetch] GET /path HTTP/1.1
		Host: 127.0.0.1:51118
		User-Agent: Go-http-client/1.1
		Accept-Encoding: gzip

		2020/06/30 10:31:59 [Fetch] HTTP/1.1 200 OK
		Content-Length: 2
		Content-Type: text/plain; charset=utf-8
		Date: Tue, 30 Jun 2020 02:31:59 GMT

		ok
		2020/06/30 10:31:59 [Fetch] GET /path3 HTTP/1.1
		Host: 127.0.0.1:51118
		User-Agent: Go-http-client/1.1
		Accept-Encoding: gzip

		2020/06/30 10:31:59 [Fetch] HTTP/1.1 200 OK
		Content-Length: 2
		Content-Type: text/plain; charset=utf-8
		Date: Tue, 30 Jun 2020 02:31:59 GMT
	*/
}

// go test -v -bench . -benchmem -run BenchmarkFetch_PostJSON
func BenchmarkFetch_PostJSON(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", body.MIMEJSON)
		w.Header().Add("x-request-id", fmt.Sprintf("trace-id-%d", time.Now().UnixNano()))
		out := newTestBaseResp(nil)
		fmt.Fprint(w, out.json())
	}))

	f := fetch.New(ts.URL, fetch.Debug(true), fetch.Interceptors(
		func(ctx context.Context, req *http.Request, handler fetch.Handler) (*http.Response, []byte, error) {
			b, err := req.GetBody()
			if b != nil {
				bs := make([]byte, req.ContentLength)
				_, err = b.Read(bs)
				log.Printf("GetBody1 read body:%s, err:%v", bs, err)
			}

			util.ResetBody(req, []byte(`{"a":1}`)) // update
			return handler(ctx, req)
		},
		// fetch.LogInterceptor 会输出以下日志内容
		fetch.DefaultLogInterceptor,
		func(ctx context.Context, req *http.Request, httpHandler fetch.Handler) (*http.Response, []byte, error) {
			b, err := req.GetBody()
			if b != nil {
				bs := make([]byte, req.ContentLength)
				_, err = b.Read(bs)
				log.Printf("GetBody2 read body:%s, err:%v", bs, err)
			}
			return httpHandler(ctx, req)
		},
	))
	for i := 0; i < b.N; i++ {
		var res testBaseResp
		ctx := context.Background()
		err := f.Post(ctx, "api/user").
			JSON(map[string]interface{}{
				"name": "ming.liu",
				"age":  18,
			}).BindJSON(&res)
		if err != nil {
			log.Printf("BenchmarkFetch_PostJSON failed. err:%v", err)
			return
		}
		log.Printf("BenchmarkFetch_PostJSON res:%+v", res)
	}
}
