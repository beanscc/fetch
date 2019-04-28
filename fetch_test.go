package fetch

import (
	"testing"
)

func Test_Fetch(t *testing.T) {
	f := New("http://www.dianping.com")
	resp := f.Get("/bar/search").Debug(true).
		Query("cityId", "2").
		Send(nil)
	out, err := resp.Body()
	t.Logf("err=%v, out=%#v", err, string(out))

	// AddHeader("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36`).
}
