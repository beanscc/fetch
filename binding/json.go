package binding

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Json bind json
type Json struct{}

// Name name
func (j *Json) Name() string {
	return "json"
}

// Bind json bind
func (j *Json) Bind(resp *http.Response, out interface{}) error {
	if resp == nil {
		return errors.New("nil resp")
	}

	if resp.Body == nil {
		return errors.New("nil resp.Body")
	}

	if err := decodeJson(resp.Body, out); err != nil {
		return err
	}

	return nil
}

// BindBody bind body
func (j *Json) BindBody(body []byte, out interface{}) error {
	return decodeJson(bytes.NewReader(body), out)
}

func decodeJson(r io.Reader, out interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(out)
}
