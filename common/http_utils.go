package common

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// HttpResp represents an HTTP response
type HttpResp[T any] struct {
	StatusCode int
	Body       T
	Headers    http.Header
}

type HttpCallParams[T any] struct {
	Url     string
	Method  string
	Headers map[string]string
	Target  *T
	Body    interface{}
}

// HttpCall makes an HTTP request
func HttpCall[T any](params HttpCallParams[T]) (*HttpResp[T], error) {
	// Create a new HTTP request
	var bodyBytes []byte
	if params.Body != nil {
		var err error
		bodyBytes, err = json.Marshal(params.Body)
		if err != nil {
			return nil, err
		}
	}

	httpReq, err := http.NewRequest(params.Method, params.Url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range params.Headers {
		httpReq.Header.Set(key, value)
	}

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response body
	err = json.Unmarshal(responseBody, params.Target)
	if err != nil {
		return nil, err
	}
	return &HttpResp[T]{
		StatusCode: resp.StatusCode,
		Body:       *params.Target,
		Headers:    resp.Header,
	}, nil
}

func BearerHeader(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}
