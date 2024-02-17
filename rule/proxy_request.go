package rule

import (
	"bytes"
	"io"
	"net/http"
)

type ProxyRequest struct {
	// Do not read the request body directly. Use Body() to read it.
	Request  *http.Request
	readBody []byte
}

func NewProxyRequest(r *http.Request) *ProxyRequest {
	return &ProxyRequest{
		Request: r,
	}
}

func (r *ProxyRequest) Body() ([]byte, error) {
	if r.readBody != nil {
		return r.readBody, nil
	}
	defer r.Request.Body.Close()
	rawBody, err := io.ReadAll(r.Request.Body)
	if err != nil {
		return nil, err
	}
	r.readBody = rawBody
	r.Request.Body = io.NopCloser(bytes.NewBuffer(r.readBody))
	return r.readBody, nil
}
