package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/beanscc/fetch/body"
	"github.com/beanscc/fetch/util"
)

func testFilterOk(ctx context.Context, req *http.Request, handler Handler) (*http.Response, []byte, error) {
	log.Printf("[filterOK] start")
	resp, bb, err := handler(ctx, req)
	if err != nil {
		log.Printf("[filterOK] err=%v", err)
		return resp, bb, err
	}

	if resp.StatusCode == http.StatusOK {
		log.Printf("[filterOK] ok")
		// return nil, errors.New("[filterOk] filtered")
		var b []byte
		b, resp.Body, err = util.DrainBody(resp.Body)
		log.Printf("[filterOk] resp.Body=%s..., err=%v", b, err)
	}

	log.Printf("[filterOK] end")
	return resp, bb, err
}

func testFilter1(ctx context.Context, req *http.Request, handler Handler) (*http.Response, []byte, error) {
	log.Printf("[filter-1] start")
	req.Header.Add("x-request-id", "xxxxx")
	var b []byte
	resp, b, err := handler(ctx, req)
	if err != nil {
		log.Printf("[filter-1] err=%v", err)
		return resp, b, err
	}

	// b, resp.Body, err = DrainBody(resp.Body)
	log.Printf("[filter-1] resp.Body=%s..., err=%v", b, err)
	log.Printf("[filter-1] end")

	return resp, b, err
}

// func retry_1(ctx context.Context, req *http.Request, handler Handler) (*http.Response, error) {
// 	log.Printf("[retry_1] start")
//
// 	var (
// 		resp *http.Response
// 		err  error
// 	)
//
// 	err = Retry(3*time.Second, 3, func(n int) error {
// 		log.Printf("[retry_1] n=%v", n)
// 		resp, err = handler(ctx, req)
// 		if err != nil || resp == nil || resp.StatusCode != 500 {
// 			if n == 2 { // 模拟第二次重试时，达到预期
// 				return nil
// 			}
// 			return errors.New("[retry_1] has err. want retry")
// 		}
//
// 		return nil
// 	})
//
// 	if err != nil {
// 		log.Printf("[retry_1] retry failed. err=%v", err)
// 		return resp, err
// 	}
// 	var b []byte
// 	b, resp.Body, err = DrainBody(resp.Body)
// 	log.Printf("[retry_1] resp.Body=%s..., err=%v", b[:100], err)
// 	log.Printf("[retry_1] end")
//
// 	return resp, err
// }

// go test -v -run Test_Fetch_Get
func Test_Fetch_Get(t *testing.T) {
	f := New("http://www.dianping.com/", Interceptors(
		testFilterOk,
		testFilter1,
		// retry_1,
	))

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

	var sr searchResp
	// err = resp.BindJSON(&sr)
	ctx := context.Background()

	err := f.Get(ctx, "/bar/search").
		//Debug(true).
		// Timeout(100*time.Millisecond).  // 超时
		Query("cityId", 2).
		Bind("json", &sr)
	t.Logf("err=%v, resp=%#v", err, sr)
}

type testBaseResp struct {
	Data interface{} `json:"data,empty"`
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
}

func newTestBaseResp(data interface{}) *testBaseResp {
	return &testBaseResp{
		Data: data,
		Code: 0,
		Msg:  "ok",
	}
}

func (r testBaseResp) String() string {
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

		fmt.Printf("bb=%s\n", bb)

		out := testBaseResp{
			Data: map[string]interface{}{
				"name": "test-post-json",
			},
		}

		w.Header().Add("X-Request-Id", r.Header.Get("X-Request-Id"))
		fmt.Fprintln(w, out.String())

		// time.Sleep(2 * time.Second)
	}))

	cUser := map[string]interface{}{
		"name":  "cc",
		"age":   18,
		"money": 10.0068,
	}

	ctx := context.Background()
	var resData map[string]interface{}
	res := newTestBaseResp(resData)
	f := New(ts.URL, Debug(false))
	f = f.WithOptions(
		// Debug(true),
		Timeout(1*time.Second),
		Interceptors(LogInterceptor(nil)),
	)
	err := f.Post(ctx, "/api/user").
		// Query("t", time.Now()).Query("nonce", "xxxxss--sss---xx").
		// // 或
		// Query("t", time.Now(), "nonce", "xxxxss--sss---xx").
		// 或
		Query("t", time.Now(), map[string]interface{}{"nonce": "xxxxss--sss---xx"}).
		JSON(cUser).
		// Form(cUserMap).
		// MultipartForm(cUserMap, fs...).
		// Timeout(10 * time.Microsecond).
		BindJSON(res)
	t.Logf("err=%v, resp=%#v", err, res)

	cUserMap := map[string]interface{}{
		"name": "cc",
		"age":  18,
	}

	filePath := "testdata/f1.txt"
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("readFile failed. err=%v", err)
	}
	fs := []body.File{
		{
			Field:   "file_1",
			Path:    filePath,
			Content: fileContent,
		},
		{
			Field:       "file_2",
			Path:        filePath,
			ContentType: "application/octet-stream",
			Content:     fileContent,
		},
	}

	rb, err := f.Post(ctx, "/api/user").
		MultipartForm(cUserMap, fs...).
		Bytes()
	t.Logf("form-data upload file. resp=%s, err=%v", rb, err)

	// cUserStr := `{"name": "cc", "age": 18}`
}
