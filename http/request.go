package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HTTP request methods
const (
	MethodGET = "GET"
)

// Response is jolokia response struct
type Response struct {
	Value  interface{}
	Error  string
	Status int
}

// DoRequest sends an HTTP request and returns an HTTP response
func DoRequest(url string) (*Response, error) {
	req, err := http.NewRequest(MethodGET, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "mackerel-plugin-jolokia")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request status code is %d", res.StatusCode)
	}
	defer res.Body.Close()

	var resp Response
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&resp); err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("error: %s", resp.Error)
	}

	return &resp, nil
}
