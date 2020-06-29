package fetch_test

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

	"github.com/beanscc/fetch"
	"github.com/beanscc/fetch/body"
	"github.com/beanscc/fetch/util"
)

func testFilterOk(ctx context.Context, req *http.Request, handler fetch.Handler) (*http.Response, []byte, error) {
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

func testFilter1(ctx context.Context, req *http.Request, handler fetch.Handler) (*http.Response, []byte, error) {
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

// go test -v -run TestFetchGet
func TestFetchGet(t *testing.T) {
	type Resp struct {
		Name   string `json:"name"`
		Age    uint8  `json:"age"`
		Addr   string `json:"address"`
		Mobile string `json:"mobile"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out := testBaseResp{
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
		w.Write(res)
	}))

	f := fetch.New(ts.URL,
		fetch.Debug(true),
		fetch.Interceptors(
			// fetch.LogInterceptor 会输出以下日志内容
			/*
				2020/06/29 23:53:26 [Fetch] method: GET, url: http://127.0.0.1:65170/api/user?id=10&name=liu, header: map[X-Request-Id:[xxxx-xxxx-xxxx]], body: , latency: 1.128272ms, status: 200, resp: {"data":{"name":"ming.liu","age":20,"address":"beijing wangfujing street","mobile":"+86-13800000000"},"code":0,"msg":"ok"}, err: <nil>, extra k1:v1
			*/
			fetch.LogInterceptor(&fetch.LogInterceptorRequest{
				ExcludeReqHeader: nil,
				MaxReqBody:       0,
				MaxRespBody:      0,
				Logger: func(ctx context.Context, format string, args ...interface{}) {
					v1, _ := ctx.Value("k1").(string)
					log.Printf(format+", extra k1:%v", append(args, v1)...)
				},
			}),
		))

	var data Resp
	tRes := newTestBaseResp(&data)
	ctx := context.WithValue(context.Background(), "k1", "v1")
	err := f.
		Get(ctx, "/api/user").
		Query("id", 10, map[string]interface{}{"name": "liu"}).
		AddHeader("x-request-id", "xxxx-xxxx-xxxx").
		// Bind("json", &tRes)
		// 或
		BindJSON(&tRes)
	if err != nil {
		t.Errorf("Test_Fetch_Get failed. err:%v", err)
		return
	}
	t.Logf("resp.data=%#v", data)

	// output:
	/*
			2020/06/29 23:48:01 [Fetch-Debug] GET /api/user?id=10&name=liu HTTP/1.1
			Host: 127.0.0.1:65110
			User-Agent: Go-http-client/1.1
			X-Request-Id: xxxx-xxxx-xxxx
			Accept-Encoding: gzip

			2020/06/29 23:48:01 [Fetch-Debug] HTTP/1.1 200 OK
			Content-Length: 122
			Content-Type: application/json
			Date: Mon, 29 Jun 2020 15:48:01 GMT

			{"data":{"name":"ming.liu","age":20,"address":"beijing wangfujing street","mobile":"+86-13800000000"},"code":0,"msg":"ok"}

		fetch_test.go:102: resp.data=fetch_test.Resp{Name:"ming.liu", Age:0x14, Addr:"beijing wangfujing street", Mobile:"+86-13800000000"}
	*/
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
	f := fetch.New(ts.URL)
	f = f.WithOptions(
		fetch.Debug(true),
		fetch.Timeout(1*time.Second),
		fetch.Interceptors(fetch.LogInterceptor(nil)),
	)
	err := f.Post(ctx, "/api/user").
		// Query("t", time.Now()).Query("nonce", "xxxxss--sss---xx").
		// // 或
		// Query("t", time.Now(), "nonce", "xxxxss--sss---xx").
		// 或
		Query("t", time.Now().Unix(), map[string]interface{}{"nonce": "xxxxss--sss---xx"}).
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
		{
			Field:       "file_2",
			Path:        filePath,
			ContentType: "application/octet-stream",
			Content:     fileContent,
		},
	}

	rb, err := f.Post(ctx, "/api/user").
		MultipartForm(cUserMap, fs...).
		AddHeader("token", "xxx-xxxx").
		AddHeader("x-request-id", "wwww-wwww").
		Bytes()
	t.Logf("form-data upload file. resp=%s, err=%v", rb, err)

	resDo := f.Post(ctx, "/api/user").
		MultipartForm(cUserMap, fs...).
		AddHeader("token", "2222xxx-xxxx").
		AddHeader("x-request-id", "22222wwww-wwww").
		Do()

	rb2, err2 := resDo.Bytes()
	t.Logf("form-data2 upload file. resp=%s, err=%v", rb2, err2)
}
