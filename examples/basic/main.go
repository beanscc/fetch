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
	// err := fetch.Get(context.Background(), ts.URL).
	err := fetch.New(ts.URL, fetch.Debug(true)).Get(context.Background(), "api/user").
		Query("id", 10).
		BindJSON(&res)
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
