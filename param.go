package listenerutil

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// ParseBodyParam 按json格式解析请求体body
func ParseBodyParam(r *http.Request, param interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	// Reset resp.body so it can be use again
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	err = json.Unmarshal(body, param)
	if err != nil {
		return err
	}
	return nil
}
