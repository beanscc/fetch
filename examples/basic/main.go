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

	f := fetch.New("https://www.dianping.com",
		fetch.Debug(false),
		// fetch.Interceptors(interceptorLog),
		fetch.Interceptors(fetch.LogInterceptor(nil)),
		fetch.Timeout(3*time.Second),
	)
	err := f.Get(context.Background(), "/bar/search").
		// Timeout(100*time.Millisecond).  // 超时
		Query("cityId", 2).
		SetHeader("k1", "123").
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
