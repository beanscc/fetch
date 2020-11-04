package fetch

import (
	"log"
	"net/http"
	"net/http/httputil"
)

func dumpRequest(req *http.Request, body bool) error {
	dump, err := httputil.DumpRequestOut(req, body)
	if err != nil {
		log.Printf("[Fetch] dump request failed. err:%v", err)
		return err
	}

	log.Printf("[Fetch] %s", dump)
	return nil
}

func dumpResponse(resp *http.Response, body bool) error {
	dump, err := httputil.DumpResponse(resp, body)
	if err != nil {
		log.Printf("[Fetch] dump response failed. err:%v", err)
		return err
	}

	log.Printf("[Fetch] %s", dump)
	return nil
}
