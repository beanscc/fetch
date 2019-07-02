# fetch
http 网络请求封装

## todo
- 支持自定义 log 打印请求和响应日志
- 完善 test
- 完善 example
- 完善 文档

## 设置服务基础域名地址

```go
func New(baseUrl string) *Fetch {}
```

## 设置请求方法和接口路径

```go
// Get
func Get(ctx context.Context, path string) *Fetch {}

// Post
func Post(ctx context.Context, path string) *Fetch {}

// Put
func Put(ctx context.Context, path string) *Fetch {}

// Delete
func Delete(ctx context.Context, path string) *Fetch {}

// Head
func Head(ctx context.Context, path string) *Fetch {}

// Option
func Option(ctx context.Context, path string) *Fetch {}
```

## 请求 path 参数设置

```go
// path 所有参数都存储在 map[string]string 这个数据结构中

// Query 加单个参数
func Query(key, value string) *Fetch {}

// 一次设置多个参数
func QueryMany(p map[string]string) *Fetch {}
```

## 请求 header 参数设置

所有 header 参数都存储在 http.Header 结构中

```go
func AddHeader(key, value string) *Fetch {}

func SetHeader(key, value string) *Fetch {}
```

## 请求 body 参数设置

所有的 body 参数都以 io.Reader 接口形式传入，各不同类型的body消息体，实现一个接口用来获取 body 的 io.Reader 接口数据

### json content-type json

### form 表单 content-type form-data

### form urlencoded content-type x-www-form-urlencoded

### 发送文件

## Interceptor 拦截器

拦截器有 2 个拦截点
- 发送 http 请求前，对请求进行拦截，可对请求数据进行预处理
- 请求发送后，对响应进行拦截，可对响应进行一些处理

多个拦截器是如何运作的？
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
b, resp.Body, err = DrainBody(resp.Body)
```
 
 
拦截器可以做什么？

通过拦截器可以做很多事，如
- 记录每次请求及响应数据的log信息
- 在请求前对参数进行签名
- 请求前对参数/body消息体进行加密，响应后对消息体进行加解密
- 自定义请求进行重试，按需自定义何种情况/多少时间间隔/重试多少次
- 对请求进行打点上报请求质量状况

请注意拦截器执行的顺序流程，合理安排多个拦截器之间的顺序关系


如何注册拦截器？

todo


## 重试机制

重试机制，通过 Interceptor 来实现

首先，自定义关于重试的 Interceptor
然后，向 Fetch 对象注册 retry 拦截器 

```go
func retry_1(ctx context.Context, req *http.Request, handler Handler) (*http.Response, error) {
	log.Printf("[retry_1] start")

	var (
		resp *http.Response
		err  error
	)

	err = Retry(3*time.Second, 3, func(n int) error {
		log.Printf("[retry_1] n=%v", n)
		resp, err = handler(ctx, req)
		if err != nil || resp == nil || resp.StatusCode != 500 {
			if n == 2 { // 模拟第二次重试时，达到预期
				return nil
			}
			return errors.New("[retry_1] has err. want retry")
		}

		return nil
	})

	if err != nil {
		log.Printf("[retry_1] retry failed. err=%v", err)
		return resp, err
	}
	var b []byte
	b, resp.Body, err = DrainBody(resp.Body)
	log.Printf("[retry_1] resp.Body=%s..., err=%v", b[:100], err)
	log.Printf("[retry_1] end")

	return resp, err
}
```

## 超时机制

```go
// 超时有 2 中方式

// 方式1. 通过 ctx
func WithContext(ctx context.Context) *Fetch {}

// 方式2. 通过 timeout
func Timeout(t time.Duration) *Fetch {}
```

## Auth 认证

```go
func BasicAuth(user, passwd string) *Fetch {}
```

## 执行 Do

```go
func Do() (*response, error) {}
```

## response 解析

支持自定义对请求响应的解析

`SetBind("json", &binding.JSON{})`

以注册的 json 解析器去解析响应 `Bind("json", &out)` 或 `BindJSON(&out)`


```go
var a someType
if err := resp.BindJSON(&a); err != ni {
	// handle err log
	return err
}
```