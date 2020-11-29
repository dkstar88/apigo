package models

import (
	"net/http"
	"net/url"
)

// APIRequest api request containing necessary data to initiate a request
type APIRequest struct {
	Method      string
	URL         string
	Arguments   url.Values
	Body        string
	ContentType string
	Headers     http.Header
}

// APIResponse holds response from an api request
type APIResponse struct {
	Body        []byte
	Status      int
	ContentType string
	Headers     http.Header
}

// NewAPIRequest creates a new APIRequest
func NewAPIRequest(method string, url string, arguments url.Values) APIRequest {
	req := APIRequest{
		Method:      method,
		URL:         url,
		Body:        "",
		ContentType: "application/json; charset=UTF-8",
		Arguments:   arguments,
		Headers:     make(map[string][]string),
	}
	return req
}
