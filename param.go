package listen

import (
	"net/http"
	"io/ioutil"
	"bytes"
	"encoding/json"
)

/*
parse json from body
 */
func ParseBodyParam(r *http.Request, param interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	// Reset resp.Body so it can be use again
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	err = json.Unmarshal(body, param)
	if err != nil {
		return err
	}
	return nil
}
