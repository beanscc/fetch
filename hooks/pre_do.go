package hooks

import (
	"net/http"
)

// PreDoer before client.Do
type PreDoer interface {
	PreDo(req *http.Request) (*http.Request, error)
}
