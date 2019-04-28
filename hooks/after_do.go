package hooks

import (
	"net/http"
)

// AfterDoer after client.Do
type AfterDoer interface {
	AfterDo(resp *http.Response) (*http.Response, error)
}
