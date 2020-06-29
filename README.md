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
		- [发送 application/json 数据](#发送-application/json-数据)
		- [发送 application/xml 数据](#发送-application/xml-数据)
		- [发送 application/x-www-form-urlencoded 表单数据](#发送-application/x-www-form-urlencoded-表单数据)
		- [发送 multipart/form-data 表单数据](#发送-multipart/form-data-表单数据)
	- [响应解析](#响应解析)
		- [如何设置自定义解析器？](#如何设置自定义解析器？)
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
		Code int         `json:"code"`
		Msg  string      `json:"msg"`
		Data interface{} `json:"data"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out := baseResp{
			Code: 0,
			Msg:  "ok",
			Data: map[string]interface{}{
				"name":   "ming.liu",
				"age":    20,
				"addr":   "beijing wangfujing street",
				"mobile": "+86-13800000000",
			},
		}

		res, _ := json.Marshal(out)
		w.Header().Set("content-type", body.MIMEJSON)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(res)
	}))

	var res Resp
	err := fetch.Get(context.Background(), ts.URL).Query("id", 10).BindJSON(&res)
	if err != nil {
		log.Printf("fetch.Get() failed. err:%v", err)
		return
	}
	log.Printf("fetch.Get() got:%+v", res) // output: fetch.Get() got:{Code:0 Msg:ok Data:map[addr:beijing wangfujing street age:20 mobile:+86-13800000000 name:ming.liu]}
}

type baseResp struct {
	Data interface{} `json:"data,empty"`
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
}
```

## API Examples

### 创建 Fetch

```go
f := fetch.New("") // 不指定 Fetch 客户端请求的基础域名，需要在 Get/Post... 等方法时，使用绝对地址

f2 := fetch.New("http://api.domain.com") // 指定基础域名地址， 后面 Get/Post ... 等方法调用时，可使用相对地址，也可以使用绝对地址

f3 := fetch.New("http://api.domain.com", fetch.Debug(true)) // 指定基础域名，同时开启 debug 模式（debug 模式，将使用标准包 "log" 以文本格式，输出请求和响应的详细日志）
```

#### Options 的使用

##### Debug

debug 默认是 false 关闭状态，若设置为 true，则为开启状态。

debug 开启状态下，会以标准包 `log` 形式输出请求和响应的信息

```go
f := fetch.New("http://api.domain.com/", Debug(true)).Get(context.Background(), "api/user").
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
// 方式1. 通过 Timeout 全局超时
f := fetch.New("", Timeout(10 * time.Second))
// 或
f = f.WithOptions(Timeout(10 * time.Second))
```

> 超时控制见：[Timeout 超时控制](#timeout-超时控制)

### Method 设置

```go
// 使用默认 Fetch
f := fetch.Get(context.Background(), "api/user")

// 自定义 Fetch 的基础域名，同时通过 option 设置每次请求的超时时间
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
f.Get(ctx, "city")  // 需要接收 Get() 返回的新 *Fetch 对象，下面的操作才不会出错
b, err := f.Query("id", "1").Text() // err != nil, err="fetch: empty method"

// 正确的方式：将下面的操作合并过来组成一个链式操作 
b, err := f.Get(ctx, "city").Query("id", "1").Text()
// 或
f = f.Get(ctx, "city")
b, err := f.Query("id", 1).Text()
```

### Query 参数设置

```go
f = f.Get(ctx, "api/user")

// query 参数可以以 key/val 成对形式设置（必须成对）
f = f.Query("id", 1).Query("type", 2).Query("name", "test")


// 或通过 map[string]interface{} 一次设置多个key/val对
f1 := f.Query(map[string]interface{}{
    "id":   1,
    "type": 2,
    "name": "test",
})

// 或者组合key/val 对形式和 map[string]interface{} 形式设置
f2 := f.Query("id", 1, map[string]interface{}{
    "type": 2,
    "name": "test",
})
```

### Header 参数设置

```go
f = f.Get(ctx, "city")

f.AddHeader("token", "token:a30e1a40-39d6-47ec-aac8-19f6258d8718").
	AddHeader("app-id", "fetch-v-4")

f.SetHeader("app-id", "fetch-v-5") // 将覆盖上面 "app-id" 的值为 "fetch-v-5"
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
f := f.Post(ctx, "user")

// 支持 string 类型 json 字符串
f.JSON(`{"name": "alice", "age": 12}`)

// 支持 []byte 类型 json
f.JSON([]byte(`{"name": "alice", "age": 12}`))

// 非 string / []byte 类型，将都调用 json.Marshal 进行序列化
f.JSON(map[string]interface{}{
	"name": "alice",
	"age": 12,
})

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

user := User{Name:"alice", Age:12}
f.JSON(user)
```

#### 发送 application/xml 数据

Content-Type: "application/xml"

xml 数据发送和 json 数据支持格式一样

```go
f := f.Post(ctx, "user")

xmlStr :=`
<note>
    <to>George</to>
    <from>John</from>
    <heading>Reminder</heading>
    <body>Don't forget the meeting!</body>
</note>
`

f.XML(xmlStr)
```

#### 发送 application/x-www-form-urlencoded 表单数据

Content-Type: "application/x-www-form-urlencoded"

```go
f := f.Post(ctx, "user")

f.Form(map[string]interface{} {
	"name": "alice",
	"age": 12,
})
```

#### 发送 multipart/form-data 表单数据

Content-Type: "multipart/form-data"

一般需要上传文件时，才使用这种类型

```go
f := f.Post(ctx, "user")


filePath := "testdata/f1.txt"
fileContent, err := ioutil.ReadFile(filePath)
if err != nil {
	log.Fatalf("readFile failed. err=%v", err)
}
f.Form(map[string]interface{}{
	"name": "alice",
	"age": 12,
}, []body.File{
	{Field: "file1", Path: filePath, Content:fileContent}, // 若未指定文件的 content-type 则，使用 http.DetectContentType(fileContent) 识别的类型，否则，使用指定的类型
	{Field: "file2", Path: filePath, ContentType: "application/octet-stream",Content:fileContent},
})
```

### 响应解析

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

// 以其他 自定义 格式解析，若以这种方式解析，请确认向 Fetch 对象注册过该 bind-type 对应的解析 func
err := f.Bind("custom-bind-type", resCustomBindRes)
```

`fetch.New()` 创建的 Fetch 对象已注册了默认的 `json` 和 `xml` 格式解析函数

#### 如何设置自定义解析器？

```go
f := fetch.New("http://www.dianping.com")

// 设置 单个 解析器
f.SetBind("my-json",  myJSONBind) // myJSONBind 必须实现 github.com/beanscc/fetch/binding.Binding 接口

// 一次设置多个解析器
f.SetBinds(map[string]binding.Binding{
	"json": &binding.JSON{},
	"xml": &binding.XML{}
})
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

拦截器有 2 个拦截点
- 发送 http 请求前，对请求进行拦截，可对请求数据进行预处理
- 请求发送后，对响应进行拦截，可对响应进行预处理

多个拦截器的执行顺序是什么？
先来看一下关于拦截器的定义

```go
// Handler 执行 http 请求
type Handler func(ctx context.Context, req *http.Request) (*http.Response, error)

// Interceptor 请求拦截器
// 多个 interceptor one,two,three 则执行顺序是 one,two,three 的 handler 调用前的执行流，然后是 handler, 接着是 three,two,one 中 handler 调用之后的执行流
type Interceptor func(ctx context.Context, req *http.Request, handler Handler) (*http.Response, error)

// ChainInterceptor 将多个 Interceptor 合并为一个
func ChainInterceptor(interceptors ...Interceptor) Interceptor {
	n := len(interceptors)

	if n > 1 {
		lastI := n - 1
		return func(ctx context.Context, req *http.Request, handler Handler) (*http.Response, error) {
			var (
				chainHandler Handler
				curI         int
			)

			chainHandler = func(currentCtx context.Context, currentReq *http.Request) (*http.Response, error) {
				if curI == lastI {
					return handler(currentCtx, currentReq)
				}
				curI++
				resp, err := interceptors[curI](currentCtx, currentReq, chainHandler)
				curI--
				return resp, err
			}

			return interceptors[0](ctx, req, chainHandler)
		}
	}

	if n == 1 {
		return interceptors[0]
	}

	// n == 0; Dummy interceptor maintained for backward compatibility to avoid returning nil.
	return func(ctx context.Context, req *http.Request, handler Handler) (*http.Response, error) {
		return handler(ctx, req)
	}
}
```

多个拦截器在执行时，首先会合并为一个拦截器（`通过函数 ChainInterceptor() 可以将多个拦截器合并成一个`）

合并后的拦截器的执行流是怎样的呢？

首先，拦截器方法中 `handler` 回调函数，就是执行 http 请求的部分，在拦截器中，handler 之前的流程被认为是拦截器的第一个拦截点，handler 之后的执行流，被认为是第二个拦截点

那么按照上面合并后的执行流，若有多个 interceptor one,two,three 则执行顺序是 one,two,three 的 handler 调用前的执行流，然后是 handler, 紧接着是 three,two,one 中 handler 调用之后的执行流

> 注意：有 2 个点需要注意 
> - 在 handler 之前对请求数据进行处理时，若需要读取 req.Body , 请使用 `req.GetBody()` 方法进行读取，否则，前面的拦截器将 req.Body 数据读取后，后面的拦截器就无法读取请求body体数据了，发送 http 请求时，也将丢失 body 消息体
> - 在 handler 之后对请求响应数据进行处理时，若需要读取 resp.Body，请使用 `DrainBody()` 函数拷贝并重置 resp.Body，否则和读取请求body 消息体一样，后面将无法从响应body中读取body消息体

拦截器可以做什么？
- 记录每次请求及响应数据的日志信息
- 在请求前对参数进行签名
- 请求前对参数/body消息体进行加密，响应后对消息体进行加解密
- 自定义请求进行重试，按需自定义何种情况/多少时间间隔/重试多少次
- 对请求进行打点上报请求质量状况

请注意拦截器执行的顺序流程，合理安排多个拦截器之间的顺序关系

如何注册拦截器？

```go
// 通过 Interceptors 设置拦截器
f := New("", Interceptors(filter1))
```

## Auth 认证

```go
func BasicAuth(user, passwd string) *Fetch {}
```