package fetch

import (
	"log"
	"net/http"
	"net/http/httputil"
)

func debugRequest(req *http.Request, body bool) error {
	dump, err := httputil.DumpRequestOut(req, body)
	if err != nil {
		log.Printf("[Fetch-Debug] Dump request failed. err=%v", err)
		return err
	}

	log.Printf("[Fetch-Debug] %s", dump)
	return nil
}

func debugResponse(resp *http.Response, body bool) error {
	dump, err := httputil.DumpResponse(resp, body)
	if err != nil {
		log.Printf("[Fetch-Debug] Dump response failed. err=%v", err)
		return err
	}

	log.Printf("[Fetch-Debug] %s", dump)
	return nil
}
