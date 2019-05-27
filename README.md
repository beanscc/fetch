# fetch
http 网络请求封装


## todo

- log 打印请求和响应日志
- timeout
- retry 支持何种情况下 retry

## 设置服务域名地址

```go
func New(host string) *Fetch {}
```

## 设置请求方法和接口路径

```go
// Get
func Get(path string) *Fetch {}

// Post
func Post(path string) *Fetch {}

// Put
func Put(path string) *Fetch {}

// Delete
func Delete(path string) *Fetch {}

// Head
func Head(path string) *Fetch {}

// Option
func Option(path string) *Fetch {}
```

## 请求 path 参数设置

```go
// path 所有参数都存储在 map[string]string 这个数据结构中

// Query 加单个参数
func Query(key, value string) *Fetch {}

// 一次设置多个参数
func QueryMap(p map[string]string) *Fetch {}
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

## 重试机制

```go
// 重试
func Retry(n int, t time.Duration, callback func()) *Fetch {}
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

```go
var a someType
if err := resp.BindJSON(&a); err != ni {
	// handle err log
	return err
}
```