package fetch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func filterOk(ctx context.Context, req *http.Request, handler Handler) (*http.Response, error) {
	log.Printf("[filterOK] start")
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("[filterOK] err=%v", err)
		return resp, err
	}

	if resp.StatusCode == http.StatusOK {
		log.Printf("[filterOK] ok")
		// return nil, errors.New("[filterOk] filtered")
		var b []byte
		b, resp.Body, err = DrainBody(resp.Body)
		log.Printf("[filterOk] resp.Body=%s..., err=%v", b, err)
	}

	log.Printf("[filterOK] end")
	return resp, err
}

func filter1(ctx context.Context, req *http.Request, handler Handler) (*http.Response, error) {
	log.Printf("[filter-1] start")
	req.Header.Add("x-request-id", "xxxxx")
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("[filter-1] err=%v", err)
		return resp, err
	}
	var b []byte
	b, resp.Body, err = DrainBody(resp.Body)
	log.Printf("[filter-1] resp.Body=%s..., err=%v", b, err)
	log.Printf("[filter-1] end")

	return resp, err
}

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

// go test -v -run Test_Fetch_Get
func Test_Fetch_Get(t *testing.T) {
	f := New("http://www.dianping.com/")
	// f.UseInterceptor(filterOk, filter1)
	f.RegisterInterceptors(
		InterceptorHandler{Name: "filterOk", Interceptor: filterOk},
		InterceptorHandler{Name: "filter1", Interceptor: filter1},
		InterceptorHandler{Name: "filter1", Interceptor: retry_1},
	)
	resp := f.Get("/bar/search").
		Debug(true).
		// Timeout(100*time.Millisecond).  // 超时
		Query("cityId", "2").
		Do()
	_, err := resp.Body()
	// out, err := resp.Body()
	// t.Logf("err=%v, out=%#v", err, string(out))

	type searchResp struct {
		List []struct {
			Value struct {
				SubTag          string `json:"subtag"`
				Location        string `json:"location"`
				MainCategoryIDS string `json:"maincategoryids"`
				DataType        string `json:"datatype"`
				ID              int    `json:"id_,string"`
				KeyWord         string `json:"suggestKeyWord"`
			} `json:"valueMap"`
		} `json:"recordList"`
		Code int `json:"code"`
	}

	var sr searchResp
	err = resp.BindJson(&sr)
	// t.Logf("err=%v", err)
	t.Logf("err=%v, resp=%#v", err, sr)
}

type baseResp struct {
	Data interface{} `json:"data"`
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
}

func (r baseResp) String() string {
	s, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return string(s)
}

func Test_Fetch_POST_JSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bb, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(fmt.Sprintf("ioutil.ReadAll err. err=%v", err))
		}

		out := baseResp{
			Code: 0,
			Msg:  "",
			Data: string(bb),
		}

		fmt.Fprintln(w, out.String()+"end")
	}))

	f := New(ts.URL)

	f.RegisterInterceptors(
		InterceptorHandler{Name: "filterOk", Interceptor: filterOk},
		InterceptorHandler{Name: "filter1", Interceptor: filter1},
	// InterceptorHandler{Name: "filter1", Interceptor: retry_1},
	)

	// cUser := map[string]interface{}{
	// 	"name": "cc",
	// 	"age":  18,
	// }

	cUserMap := map[string]string{
		"name": "cc",
		"age":  "18",
	}

	// cUserStr := `{"name": "cc", "age": 18}`

	resp := f.Post("/api/v1/user").
		Debug(true).
		Query("t", time.Now().String()).
		Query("nonce", "xxxxss--sss---xx").
		// JSON(cUser).
		// Form(cUserMap).
		MultipartForm(cUserMap).
		Do()

	_, err := resp.Body()

	var sr baseResp
	err = resp.BindJson(&sr)
	// t.Logf("err=%v", err)
	t.Logf("err=%v, resp=%#v", err, sr)
}
