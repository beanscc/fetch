package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// JSON bind json
type JSON struct{}

// Name name
func (j JSON) Name() string {
	return "json"
}

// Bind json bind
func (j *JSON) Bind(resp *http.Response, body []byte, out interface{}) error {
	if resp == nil {
		return errors.New("fetch.binding.JSON: nil resp")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch.binding.JSON: incorrect response status code(%v)", resp.StatusCode)
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("fetch.binding.JSON: %v", err)
	}

	return nil
}
