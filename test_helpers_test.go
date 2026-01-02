package evateamclient

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

var testCtx = context.Background()

// mockHTTPClient is a simple mock for HTTPClient interface
type mockHTTPClient struct {
	response  *req.Response
	err       error
	urlCheck  func(string) bool
	bodyCheck func([]byte) bool
}

func (m *mockHTTPClient) Post(ctx context.Context, body []byte, url string) (*req.Response, error) {
	if m.urlCheck != nil && !m.urlCheck(url) {
		return nil, fmt.Errorf("request url does not match")
	}
	if m.bodyCheck != nil && !m.bodyCheck(body) {
		return nil, fmt.Errorf("request body does not match")
	}

	return m.response, m.err
}

func newTestClient(t *testing.T) (*Client, *mockHTTPClient) {
	cfg := Config{
		BaseURL:  "https://api.eva.team",
		APIToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	mockHTTP := &mockHTTPClient{}
	client.httpClient = mockHTTP

	return client, mockHTTP
}

func mockResponse(statusCode int, body string) *req.Response {
	r := req.C().R()
	resp := &req.Response{
		Request: r,
		Response: &http.Response{
			StatusCode: statusCode,
		},
	}
	resp.SetBodyString(body)
	return resp
}
