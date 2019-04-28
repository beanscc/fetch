package hooks

import (
	"log"
	"net/http"
	"net/http/httputil"
)

// Debug debug
type Debug struct {
	log log.Logger
}

// PreDo before client.Do
func (d *Debug) PreDo(req *http.Request) (*http.Request, error) {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		d.log.Printf("[Fetch-Debug-Req] failed to dump http request. err: %v", err)
		return req, nil
	}

	d.log.Printf("[Fetch-Debug-Req] %s", dump)
	return req, nil
}

// AfterDo after client.Do
func (d *Debug) AfterDo(resp *http.Response) (*http.Response, error) {
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		d.log.Printf("[Fetch-Debug-Resp] failed to dump http response. err: %v", err)
		return resp, nil
	}

	d.log.Printf("[Fetch-Debug-Resp] %s", dump)
	return resp, nil
}
