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
	f := fetch.New(ts.URL, &fetch.Options{
		Debug: true,
		Interceptors: []fetch.Interceptor{
			// fetch.LogInterceptor 会输出请求和响应日志
			fetch.LogInterceptor(&fetch.LogInterceptorRequest{
				MaxReqBody: 5,
				// MaxRespBody: 10,
				Logger: func(ctx context.Context, format string, args ...interface{}) {
					v1, _ := ctx.Value("k1").(string)
					allArgs := []interface{}{v1}
					allArgs = append(allArgs, args...)
					log.Printf("extra k1:%v, "+format, allArgs...)
				},
			}),
		},
	})

	ctx := context.WithValue(context.Background(), "k1", "v1")

	var data Resp
	res := newBaseResp(&data)
	err := f.Post(ctx, "api/user").
		AddHeader("hk_1", "hk_1_val").
		AddHeader(map[string]interface{}{
			"hk_2": 24,
			"hk_3": "hk_3_val",
		}).
		AddHeader("hk_4", 4, map[string]interface{}{"hk_5": 66.66}, "hk_6", "hk_6_val").
		SetHeader("hk_1", 111).
		Query("id", 10).
		JSON(`{"age": 18}`).
		BindJSON(&res)
	if err != nil {
		log.Printf("fetch.Get() failed. err:%v", err)
		return
	}
	log.Printf("fetch.Get() data:%+v", res.Data) // output: fetch.Get() data:&{Name:ming.liu Age:20 Addr:beijing wangfujing street Mobile:+86-13800000000}

	// output:
	/*
		2020/07/01 16:41:38 [Fetch] GET /api/user?id=10 HTTP/1.1
		Host: 127.0.0.1:50305
		User-Agent: Go-http-client/1.1
		Hk_1: 111
		Hk_2: 24
		Hk_3: hk_3_val
		Hk_4: 4
		Hk_5: 66.66
		Hk_6: hk_6_val
		Accept-Encoding: gzip

		b
		{"age": 18}
		0

		2020/07/01 16:41:38 [Fetch] HTTP/1.1 200 OK
		Content-Length: 122
		Content-Type: application/json
		Date: Wed, 01 Jul 2020 08:41:38 GMT

		{"data":{"name":"ming.liu","age":20,"address":"beijing wangfujing street","mobile":"+86-13800000000"},"code":0,"msg":"ok"}
		2020/07/01 16:41:38 extra k1:v1, [Fetch] method: POST,, url: http://127.0.0.1:50305/api/user?id=10, header: map[Hk_1:[111] Hk_2:[24] Hk_3:[hk_3_val] Hk_4:[4] Hk_5:[66.66] Hk_6:[hk_6_val]], body: '{"age...', latency: 1.038371ms, status: 200, resp: {"data":{"name":"ming.liu","age":20,"address":"beijing wangfujing street","mobile":"+86-13800000000"},"code":0,"msg":"ok"}, err: <nil>
		2020/07/01 16:41:38 fetch.Get() data:&{Name:ming.liu Age:20 Addr:beijing wangfujing street Mobile:+86-13800000000}
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
