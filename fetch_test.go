package fetch

import (
	"context"
	"log"
	"net/http"
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
		log.Printf("[filterOk] resp.Body=%v..., err=%v", b[:100], err)
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

	log.Printf("[filter-1] end")

	return resp, err
}

func Test_Fetch_Get(t *testing.T) {
	f := New("http://www.dianping.com/")
	f.UseInterceptor(filterOk, filter1)
	resp := f.Get("/bar/search").Debug(true).
		Query("cityId", "2").
		Send(nil).Do()
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
	err = resp.BindJSON(&sr)
	// t.Logf("err=%v", err)
	t.Logf("err=%v, resp=%#v", err, sr)
}

func Test_Fetch_POST_JSON(t *testing.T) {
	f := New("https://1d4f258f-c87a-4333-ac92-5735c96a93f9.mock.pstmn.io")

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
		// SendJSON(cUser).
		Send(XWWWFormURLEncoded{cUserMap}).
		// SendJSONStr(cUserStr).
		Do()

	_, err := resp.Body()

	type uResp struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
		Code int `json:"code"`
	}

	var sr uResp
	err = resp.BindJSON(&sr)
	// t.Logf("err=%v", err)
	t.Logf("err=%v, resp=%#v", err, sr)
}
