# fetch
http client 网络请求封装

## Overview

涵盖功能:
- 支持 Get/Post/Put/Delete/Patch/Options/Head 等方法
- 支持自定义设置 client
- 支持 debug 模式打印请求和响应详细
- 支持 timeout 超时和 ctx 超时设置
- 支持自定义 Interceptor 拦截器设置
- 支持自定义 Bind 解析请求响应


## Contents

- [Overview](#overview)
- [Contents](#contents)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Examples](#api-examples)
	- [创建 Fetch](#创建-fetch)
		- [Options 的使用](#options-的使用)
			- [Debug Option](#debug)
			- [Timeout Option](#timeout)
	- [Method 设置](#method-设置)
	- [Query 参数设置](#query-参数设置)
	- [Header 参数设置](#header-参数设置)
	- [Body 参数设置](#body-参数设置)
		- [发送 application/json 数据](#发送-applicationjson-数据)
		- [发送 application/xml 数据](#发送-applicationxml-数据)
		- [发送 application/x-www-form-urlencoded 表单数据](#发送-applicationx-www-form-urlencoded-表单数据)
		- [发送 multipart/form-data 表单数据](#发送-multipartform-data-表单数据)
	- [Resp 响应解析](#Resp-响应解析)
		- [如何自定义解析器](#如何自定义解析器)
	- [Timeout 超时控制](#timeout-超时控制)
	- [Interceptor 拦截器](#interceptor-拦截器)

## Installation

- install fetch

```
go get -u github.com/beanscc/fetch
```

- import it in your code

```
import "github.com/beanscc/fetch"
```

## Quick Start

```go
// github.com/beanscc/fetch/examples/basic/main.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/beanscc/fetch"
	"github.com/beanscc/fetch/body"
)

func main() {
	type Resp struct {
		Name   string `json:"name"`
		Age    uint8  `json:"age"`
		Addr   string `json:"address"`
		Mobile string `json:"mobile"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out := baseResp{
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
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(res)
	}))

	// var data Resp
	// res := newBaseResp(&data)
	// err := fetch.Get(context.Background(), ts.URL+"/api/user").
	// 	Query("id", 10).
	// 	BindJSON(&res)

	// OR

	f := fetch.New(ts.URL,
		fetch.Debug(true),
		fetch.Interceptors(
			// fetch.LogInterceptor 会输出以下日志内容
			// 2020/06/30 17:34:12 extra k1:v1, [Fetch] method: GET, url: http://127.0.0.1:60574/api/user?id=10, header: map[], body: , latency: 1.146575ms, status: 200, resp: {"data":{"name":"ming.liu","age":20,"address":"beijing wangfujing street","mobile":"+86-13800000000"},"code":0,"msg":"ok"}, err: <nil>
			fetch.LogInterceptor(&fetch.LogInterceptorRequest{
				ExcludeReqHeader: nil,
				MaxReqBody:       0,
				MaxRespBody:      0,
				Logger: func(ctx context.Context, format string, args ...interface{}) {
					v1, _ := ctx.Value("k1").(string)
					allArgs := []interface{}{v1}
					allArgs = append(allArgs, args...)
					log.Printf("extra k1:%v, "+format, allArgs...)
				},
			}),
		))

	ctx := context.WithValue(context.Background(), "k1", "v1")

	var data Resp
	res := newBaseResp(&data)
	err := f.Get(ctx, "api/user").
		Query("id", 10).
		BindJSON(&res)
	if err != nil {
		log.Printf("fetch.Get() failed. err:%v", err)
		return
	}
	log.Printf("fetch.Get() data:%+v", res.Data) // output: fetch.Get() data:&{Name:ming.liu Age:20 Addr:beijing wangfujing street Mobile:+86-13800000000}

	// output:
	/*
		2020/06/30 17:34:12 [Fetch-Debug] GET /api/user?id=10 HTTP/1.1
		Host: 127.0.0.1:60472
		User-Agent: Go-http-client/1.1
		Accept-Encoding: gzip

		2020/06/30 17:34:12 [Fetch-Debug] HTTP/1.1 200 OK
		Content-Length: 122
		Content-Type: application/json
		Date: Tue, 30 Jun 2020 09:34:12 GMT

		{"data":{"name":"ming.liu","age":20,"address":"beijing wangfujing street","mobile":"+86-13800000000"},"code":0,"msg":"ok"}
		2020/06/30 17:34:12 extra k1:v1, [Fetch] method: GET, url: http://127.0.0.1:60574/api/user?id=10, header: map[], body: , latency: 1.146575ms, status: 200, resp: {"data":{"name":"ming.liu","age":20,"address":"beijing wangfujing street","mobile":"+86-13800000000"},"code":0,"msg":"ok"}, err: <nil>
		2020/06/30 17:34:12 fetch.Get() data:&{Name:ming.liu Age:20 Addr:beijing wangfujing street Mobile:+86-13800000000}
	*/
}

type baseResp struct {
	Data interface{} `json:"data,empty"`
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
}

func newBaseResp(data interface{}) *baseResp {
	return &baseResp{
		Data: data,
		Code: 0,
		Msg:  "ok",
	}
}
```

## API Examples

### 创建 Fetch

```go
// 不指定 Fetch 客户端请求的基础域名，需要在 Get/Post... 等方法时，使用绝对地址
f := fetch.New("")

// 指定基础域名地址， 后面 Get/Post ... 等方法调用时，可使用相对地址，也可以使用绝对地址
f2 := fetch.New("http://api.domain.com")

// 指定基础域名，同时开启 debug 模式（debug 模式，将使用标准包 "log" 以文本格式，输出请求和响应的详细日志）
f3 := fetch.New("http://api.domain.com", fetch.Debug(true))
```

#### Options 的使用

##### Debug

debug 默认是 false 关闭的，若设置为 true，则为开启状态。debug 开启时，将以标准包 `log` 文本形式输出请求和响应的信息

```go
f := fetch.New(ts.URL, fetch.Debug(true)).
		Get(context.Background(), "api/user").
		Query("id", 10).
		Bytes()

// output
/*
2020/06/30 01:15:55 [Fetch-Debug] GET /api/user?id=10 HTTP/1.1
Host: 127.0.0.1:49893
User-Agent: Go-http-client/1.1
Accept-Encoding: gzip

2020/06/30 01:15:55 [Fetch-Debug] HTTP/1.1 200 OK
Content-Length: 119
Content-Type: application/json
Date: Mon, 29 Jun 2020 17:15:55 GMT

{"data":{"addr":"beijing wangfujing street","age":20,"mobile":"+86-13800000000","name":"ming.liu"},"code":0,"msg":"ok"}
*/
```

##### Timeout

```go
// Timeout 设置 Fetch 全局超时
f := fetch.New("", fetch.Timeout(10*time.Second))
// 或 设置某次请求的超时时间
f = f.WithOptions(fetch.Timeout(3 * time.Second))
```

> 超时控制也可以通过 context 设置超时时间：[Timeout 超时控制](#timeout-超时控制)

### Method 设置

```go
// 使用默认 Fetch
f := fetch.Get(context.Background(), "api/user")

// 自定义 Fetch 的基础域名，同时通过 timeout option 设置每次请求的超时时间
f1 := fetch.New("http://api.domain.com/", fetch.Timeout(10 *time.Second))

f1tmp := f1.Get(context.Background, "api/user").
	Query("id",10)
...

f2tmp := f1.Post(context.Background, "api/user").
	JSON(map[string]interface{}{"id": 10, "name": "ming.liu"})
...

f3tmp := f1.Method("Get", "api/user")
...

```

每个 Method() 都将返回一个新的 `*Fetch` 对象，该对象包含原 `*Fetch` 对象属性基础选项属性（`ctx`, `req` 和 `err` 等属于一次性请求参数，不在 clone 范围内）；
所以，若在 Method() 方法后，进行非链式操作，必须接收 Method() 方法或其链式操作后返回的新 `*Fetch` 对象


```go
// 错误使用方式
f.Get(ctx, "city")                  // 需要接收 Get() 返回的新 *Fetch 对象，下面的操作才不会出错
b, err := f.Query("id", "1").Text() // err != nil, err="fetch: empty method"

// 正确的方式：将下面的操作组成一个链式操作
b, err := f.Get(ctx, "city").Query("id", "1").Text()
// 或
f = f.Get(ctx, "city")
b, err := f.Query("id", 1).Text()
```

### Query 参数设置

```go
f = f.Get(ctx, "api/user")

// query 参数以 key/val 对形式设置（必须成对）
f = f.Query("id", 1).Query("age", 12).Query("name", "ming.liu")

// 或通过 map[string]interface{} 一次设置多个key/val对
f1 := f.Query(map[string]interface{}{
	"id":   1,
	"age": 12,
})

// 或者 key/val 对和 map[string]interface{} 交替形式设置
f2 := f.Query("id", 1, map[string]interface{}{
	"age": 12,
	"name": "ming.liu",
}, "height", 175)
```

### Header 参数设置

```go
f = f.Get(ctx, "city")

f.AddHeader("app-key", "a30e1a40-39d6-47ec-aac8-19f6258d8718").
	AddHeader("app-time", 1593506501)

f.SetHeader("app-time", time.Now().UnixNano()) // 将覆盖上面 "app-time" 的值
```

### Body 参数设置


```go
// import "github.com/beanscc/fetch/body"

// Body 构造请求的body
type Body interface {
	// Body 构造http请求body
	Body() (io.Reader, error)
	// ContentType 返回 body 体结构相应的 Header content-type 类型
	ContentType() string
}

// Body 接口实现检查
var (
	_ Body = &JSON{}
	_ Body = &XML{}
	_ Body = &Form{}
	_ Body = &MultipartForm{}
)
```

#### 发送 application/json 数据

Content-Type: "application/json"

```go
f := f.Post(ctx, "api/user")

// 支持 string 类型 json 字符串
f.JSON(`{"name": "alice", "age": 12}`)

// 支持 []byte 类型 json
f.JSON([]byte(`{"name": "alice", "age": 12}`))

// 非 string / []byte 类型，将都调用 json.Marshal 进行序列化
f.JSON(map[string]interface{}{
	"name": "alice",
	"age":  12,
})

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

user := User{Name: "alice", Age: 12}
f.JSON(user)
```

> 示例：github.com/beanscc/fetch/fetch_test.go:TestFetchPostJSON

```go
func TestFetchPostJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", body.MIMEJSON)
		w.Header().Add("x-request-id", fmt.Sprintf("trace-id-%d", time.Now().UnixNano()))
		out := newTestBaseResp(nil)
		fmt.Fprintln(w, out.json())
	}))

	var res testBaseResp
	f := fetch.New(ts.URL, fetch.Debug(true), fetch.Interceptors(
		// fetch.LogInterceptor 会输出以下日志内容
		fetch.LogInterceptor(&fetch.LogInterceptorRequest{
			ExcludeReqHeader: nil,
			MaxReqBody:       0,
			MaxRespBody:      0,
			Logger: func(ctx context.Context, format string, args ...interface{}) {
				log.Printf(format, args...)
			},
		}),
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
		2020/06/30 16:09:59 [Fetch-Debug] POST /api/user HTTP/1.1
		Host: 127.0.0.1:58717
		User-Agent: Go-http-client/1.1
		Transfer-Encoding: chunked
		Content-Type: application/json
		Accept-Encoding: gzip

		1c
		{"age":18,"name":"ming.liu"}
		0

		2020/06/30 16:09:59 [Fetch-Debug] HTTP/1.1 200 OK
		Content-Length: 22
		Content-Type: application/json
		Date: Tue, 30 Jun 2020 08:09:59 GMT
		X-Request-Id: trace-id-1593504599030600000

		{"code":0,"msg":"ok"}
		2020/06/30 16:09:59 [Fetch] method: POST, url: http://127.0.0.1:60661/api/user, header: map[Content-Type:[application/json]], body: {"age":18,"name":"ming.liu"}, latency: 995.441µs, status: 200, resp: {"code":0,"msg":"ok"}
		--- PASS: TestFetchPostJSON (0.00s)
		    fetch_test.go:283: TestFetchPostJSON res:{Data:<nil> Code:0 Msg:ok}
	*/
}
```

#### 发送 application/xml 数据

Content-Type: "application/xml"

xml 数据发送和 json 数据支持格式一样

```go
f := f.Post(ctx, "api/user")

xmlStr := `
<note>
    <to>George</to>
    <from>John</from>
    <heading>Reminder</heading>
    <body>Don't forget the meeting!</body>
</note>
`

f.XML(xmlStr)
```

> 示例：github.com/beanscc/fetch/fetch_test.go:TestFetchPostXML

```go
func TestFetchPostXML(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", body.MIMEXML)
		w.Header().Add("x-request-id", fmt.Sprintf("trace-id-%d", time.Now().UnixNano()))
		out := newTestBaseResp(nil)
		fmt.Fprintln(w, out.xml())
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
		2020/06/30 16:09:05 [Fetch-Debug] POST /api/user HTTP/1.1
		Host: 127.0.0.1:58708
		User-Agent: Go-http-client/1.1
		Transfer-Encoding: chunked
		Content-Type: application/xml
		Accept-Encoding: gzip

		56
		<user id="6135200011057538"><name>si.li</name><age>20</age><height>175</height></user>
		0

		2020/06/30 16:09:05 [Fetch-Debug] HTTP/1.1 200 OK
		Content-Length: 57
		Content-Type: application/xml
		Date: Tue, 30 Jun 2020 08:09:05 GMT
		X-Request-Id: trace-id-1593504545433384000

		<testBaseResp><code>0</code><msg>ok</msg></testBaseResp>
		--- PASS: TestFetchPostXML (0.00s)
		    fetch_test.go:340: TestFetchPostXML res:{Data:<nil> Code:0 Msg:ok}
	*/
}
```

#### 发送 application/x-www-form-urlencoded 表单数据

Content-Type: "application/x-www-form-urlencoded"

```go
f := f.Post(ctx, "user")

f.Form(map[string]interface{}{
	"name": "alice",
	"age":  12,
})
```

> 示例：github.com/beanscc/fetch/fetch_test.go:TestFetchPostForm

```go
func TestFetchPostForm(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", body.MIMEJSON)
		w.Header().Add("x-request-id", fmt.Sprintf("trace-id-%d", time.Now().UnixNano()))
		out := newTestBaseResp(nil)
		fmt.Fprintln(w, out.json())
	}))

	ctx := context.Background()
	f := fetch.New(ts.URL, fetch.Debug(true))
	resBody, err := f.Post(ctx, "api/user").
		Form(map[string]interface{}{
			"name": "wang.wu",
			"age":  25,
		}).Text()
	if err != nil {
		t.Errorf("TestFetchPostForm failed. err:%v", err)
		return
	}
	t.Logf("TestFetchPostForm resp body:%s", resBody)

	// output:
	/*
		2020/06/30 16:08:06 [Fetch-Debug] POST /api/user HTTP/1.1
		Host: 127.0.0.1:58696
		User-Agent: Go-http-client/1.1
		Transfer-Encoding: chunked
		Content-Type: application/x-www-form-urlencoded
		Accept-Encoding: gzip

		13
		age=25&name=wang.wu
		0

		2020/06/30 16:08:06 [Fetch-Debug] HTTP/1.1 200 OK
		Content-Length: 22
		Content-Type: application/json
		Date: Tue, 30 Jun 2020 08:08:06 GMT
		X-Request-Id: trace-id-1593504486872096000

		{"code":0,"msg":"ok"}
		--- PASS: TestFetchPostForm (0.00s)
		    fetch_test.go:386: TestFetchPostForm resp body:{"code":0,"msg":"ok"}
	*/
}
```

#### 发送 multipart/form-data 表单数据

Content-Type: "multipart/form-data"

`可上传文件`

> 示例：github.com/beanscc/fetch/fetch_test.go:TestFetchPostMultipartForm

```go
func TestFetchPostMultipartForm(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", body.MIMEJSON)
		w.Header().Add("x-request-id", fmt.Sprintf("trace-id-%d", time.Now().UnixNano()))
		out := newTestBaseResp(nil)
		fmt.Fprintln(w, out.json())
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
		2020/06/30 16:18:38 [Fetch-Debug] POST /api/user HTTP/1.1
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

		2020/06/30 16:18:38 [Fetch-Debug] HTTP/1.1 200 OK
		Content-Length: 22
		Content-Type: application/json
		Date: Tue, 30 Jun 2020 08:18:38 GMT
		X-Request-Id: trace-id-1593505118084005000

		{"code":0,"msg":"ok"}
		--- PASS: TestFetchPostMultipartForm (0.00s)
		    fetch_test.go:365: TestFetchPostMultipartForm resp body:{"code":0,"msg":"ok"}
	*/
}
```

### Resp 响应解析

```go
f := f.Post(ctx, "api/user")

// 支持 string 类型 json 字符串
f.JSON(`{"name": "alice", "age": 12}`)

// 获取 *fetch.response
fRes, err := f.Do()

// 获取 *http.Response
res, err := f.Resp()

// 获取 http 响应 body 消息体的 []byte
resBytes, err := f.Bytes()

// 获取 http 响应 body 消息体的 string
resStr, err := f.Text()

// ==== 对响应body消息体进行结构化解析 ====

// 以 json 格式解析
err := f.BindJSON(&resJson)

// 以 xml 格式解析
err := f.BindXML(&resXml)
```

> `fetch.New()` 创建的 Fetch 对象已注册了默认的 `json` 和 `xml` 格式解析函数

#### 如何自定义解析器

```go
// 先注册解析器
// 方式1: New() 时，注册 Bind option
f := fetch.New("", fetch.Bind(map[string]binding.Binding{"custom-bind-type", customBindFn}))

// 方式2: 通过 WithOptions() 注册 Bind
f.WithOptions(fetch.Bind(map[string]binding.Binding{"custom-bind-type", customBindFn}))

// 解析使用
err := f.Bind("custom-bind-type", customBindFn)
```

### Timeout 超时控制

```go
// 方式1. 通过 Timeout 全局超时
f := fetch.New("", Timeout(10 * time.Second))
// 或
f = f.WithOptions(Timeout(10 * time.Second))

// 方式2. 通过 ctx 单次请求超时设置
ctx, cancel := context.WithTimeout(context.Background, 10 * time.Second)
defer cancel()
f = f.Get(ctx, "api/user")
```

### Interceptor 拦截器

拦截器可以做什么？
- 记录每次请求及响应数据的日志信息，可参见 `fetch.LogInterceptor()`
- 对请求进行打点上报请求质量状况
- 在请求前对参数进行签名
- 请求前对参数/body消息体进行加密，响应后对消息体进行加解密
- 自定义请求进行重试，自定义何种情况/多少时间间隔/重试多少次

> 请注意拦截器执行的顺序流程，合理安排多个拦截器之间的顺序关系

拦截点:
- 发送 http 请求前，对请求进行拦截，可对请求数据进行预处理
- 请求发送后，对响应进行拦截，可对响应进行预处理

多个拦截器的执行顺序是什么？
先来看一下关于拦截器的定义

```go
// Handler http req handle
type Handler func(ctx context.Context, req *http.Request) (*http.Response, []byte, error)

// Interceptor 请求拦截器
// 多个 interceptor one,two,three 则执行顺序是 one,two,three 的 handler 调用前的执行流，然后是 handler, 接着是 three,two,one 中 handler 调用之后的执行流
type Interceptor func(ctx context.Context, req *http.Request, httpHandler Handler) (*http.Response, []byte, error)

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
```

多个拦截器在执行时，首先会合并为一个拦截器（`通过函数 chainInterceptor() 可以将多个拦截器合并成一个`）

合并后的拦截器的执行流是怎样的呢？

首先，拦截器方法中 `handler` 回调函数，主要是执行 Client.Do()，在拦截器中，handler 之前的流程被认为是拦截器的第一个拦截点，handler 之后的执行流，被认为是第二个拦截点

那么按照上面合并后的执行流，若有多个 interceptor one,two,three 则执行顺序是 one,two,three 的 handler 调用前的执行流，然后是 handler, 紧接着是 three,two,one 中 handler 调用之后的执行流

> 注意：有 2 个点需要注意 
> - 在 handler 之前对请求数据进行处理时，若需要读取 req.Body , 请使用 `req.GetBody() 或 util.DrainBody()` 方法进行读取，否则，前面的拦截器将 req.Body 数据读取后，后面的拦截器就无法读取请求body体数据了，发送 http 请求时，也将丢失 body 消息体
> - 在 handler 之后对请求响应数据进行处理时，若需要读取 resp.Body，可使用 `util.DrainBody()` 函数拷贝并重置 resp.Body，否则和读取请求body 消息体一样，后面将无法从响应body中读取body消息体

如何注册拦截器？

```go
// 通过 Interceptors 设置拦截器
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
```