# fetch
http 网络请求封装

## Overview

涵盖功能

todo
- 支持自定义 log 打印请求和响应日志
- 完善 test
- 完善 example
- 完善 文档


## Installation

- install

```
go get -u github.com/beanscc/fetch
```

- import

```
import "github.com/beanscc/fetch"
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/beanscc/fetch"
)

func main() {
	type searchResp struct {
		List []struct {
			Value struct {
				SubTag          string `json:"subtag" xml:"subtag"`
				Location        string `json:"location"`
				MainCategoryIDS string `json:"maincategoryids"`
				DataType        string `json:"datatype"`
				ID              int    `json:"id_,string"`
				KeyWord         string `json:"suggestKeyWord"`
			} `json:"valueMap"`
		} `json:"recordList"`
		Code int `json:"code"`
	}

	var sr, sr2 searchResp

	f := fetch.New("http://www.dianping.com", fetch.Interceptors(interceptorLog), fetch.Timeout(3*time.Second))
	err := f.Get(context.Background(), "/bar/search").
		// Timeout(100*time.Millisecond).  // 超时
		Query("cityId", 2).
		BindJSON(&sr)
	fmt.Printf("err=%v, res=%v\n", err, sr)

	// 请求方法内部会 clone 一个新的 Fetch 对象
	fmt.Println("\n==============================================")

	err2 := f.Get(context.Background(), "/bar/search").
		// Timeout(100*time.Millisecond).  // 超时
		Query("cityId", 3).
		BindJSON(&sr2)
	fmt.Printf("err2=%v, res2=%v\n", err2, sr2)
}

func interceptorLog(ctx context.Context, req *http.Request, handler fetch.Handler) (*http.Response, []byte, error) {
	log.Printf("[log] start ...")
	resp, b, err := handler(ctx, req)
	if err != nil {
		log.Printf("[log] handler failed. err=%v", err)
		return resp, b, err
	}

	if resp.StatusCode == http.StatusOK {
		log.Printf("[log] http.StatusCode = ok")
		//var b []byte
		//b, resp.Body, err = fetch.DrainBody(resp.Body)
		log.Printf("[log] resp.Body=%s, err=%v", b, err)
	}

	log.Printf("[log] end ...")
	return resp, b, err
}

```

## API Examples

### 设置服务基础域名地址

```go
f := fetch.New("") // 不指定基础域名

f := fetch.New("http://www.baidu.com") // 给定基础域名

f := fetch.New("http://www.dianping.com", Debug(true)) // 给定基础域名，同时设置一些选项
```

### 设置请求方法和接口路径

```go
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
```

每个 Method() 都将返回一个新的 `*Fetch` 对象，该对象包含原 `*Fetch` 对象属性基础选项属性（`req` 和 `ctx` 等属于每次请求范围内的参数）；
所以，若在 Method() 方法后，进行非链式操作，必须使用某个变量接收 Method() 方法或其链式操作后返回的新 `*Fetch` 对象

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

## 请求 Query 参数设置

```go
f = f.Get(ctx, "city")

// query 参数一个一个设置
f.Query("id", 1).Query("type", 2).Query("name", "test")

// 或通过 map 一次设置
f.QueryMany(map[string]interface{}{
    "id":   1,
    "type": 2,
    "name": "test",
})

// 也可以组合设置
f.Query("id", 1).QueryMany(map[string]interface{}{
    "type": 2,
    "name": "test",
})
```

## 请求 Header 参数设置

```go
f = f.Get(ctx, "city")

f.AddHeader("token", "token:a30e1a40-39d6-47ec-aac8-19f6258d8718").
	AddHeader("x-request-id", "997ba960-a512-49fc-8464-8c6685aee529").
	AddHeader("app-id", "fetch-v-4")

f.SetHeader("app-id", "fetch-v-5") // 将覆盖上面 "app-id" 的值为 "fetch-v-5"
```

## 请求 Body 参数设置


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

### 发送 json 数据

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

### 发送 xml 格式数据

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

### 发送表单：application/x-www-form-urlencoded

Content-Type: "application/x-www-form-urlencoded"

```go
f := f.Post(ctx, "user")

f.Form(map[string]string{
	"name": "alice",
	"age": "12",
})
```

### 发送表单: multipart/form-data

Content-Type: "multipart/form-data"

一般需要上传文件时，才使用这种类型

```go
f := f.Post(ctx, "user")

f.Form(map[string]string{
	"name": "alice",
	"age": "12",
}, []body.File{
	{Field:"f1", Path:"f1.txt"},
})
```

## 响应解析

```go
f := f.Post(ctx, "user")

// 支持 string 类型 json 字符串
f.JSON(`{"name": "alice", "age": 12}`)

// 获取 http.Response
res, err := f.Resp()

// 获取 http 响应 body 消息体的 []byte
resBytes, err := f.Bytes()

// 获取 http 响应 body 消息体的 string
resStr, err := f.Text()

// ==== 对响应body消息体进行结构化解析 ====

// 以 json 格式解析 
err := f.BindJSON(resJson)

// 以 xml 格式解析
err := f.BindXML(resXml)

// 以其他 自定义 格式解析，若以这种方式解析，请确认向 Fetch 对象注册过该 bind-type 对应的解析 func
err := f.Bind("customer-bind-type", resCustomerBindRes)
```

`fetch.New()` 创建的 Fetch 对象已注册了默认的 `json` 和 `xml` 格式解析函数

如何设置自定义解析器？

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

## Debug

debug 默认是 false 关闭状态，若设置为 true，则为开启状态。

debug 开启状态下，会以标准包 `log` 形式输出请求和响应的信息

```go
f := fetch.New("http://www.dianping.com", Debug(true))
```

## 超时控制

```go
// 方式1. 通过 ctx 单次请求超时设置
ctx, cancel := context.WithTimeout(context.Background, 10 * time.Second)
defer cancel()
f = f.Get(ctx, "/api/user")


// 方式2. 通过 Timeout 全局超时
f := fetch.New("", Timeout(10 * time.Second))
// 或
f = f.WithOptions(Timeout(10 * time.Second))
```

## Auth 认证

```go
func BasicAuth(user, passwd string) *Fetch {}
```

## Interceptor 拦截器

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


响应重置示例

```go
var (
	b   []byte
	err error
)
// 读取响应body消息体，并重置响应body
b, resp.Body, err = fetch.DrainBody(resp.Body)
```
 
 
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