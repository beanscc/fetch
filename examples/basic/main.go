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
