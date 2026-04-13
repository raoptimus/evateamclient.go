/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

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

	client, err := NewClient(&cfg)
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

// sequentialMockHTTPClient returns different responses for sequential calls.
type sequentialMockHTTPClient struct {
	responses []*req.Response
	errors    []error
	callIdx   int
	urlCheck  func(string) bool
	bodyCheck func([]byte) bool
}

func (m *sequentialMockHTTPClient) Post(ctx context.Context, body []byte, url string) (*req.Response, error) {
	if m.urlCheck != nil && !m.urlCheck(url) {
		return nil, fmt.Errorf("request url does not match")
	}
	if m.bodyCheck != nil && !m.bodyCheck(body) {
		return nil, fmt.Errorf("request body does not match")
	}

	idx := m.callIdx
	m.callIdx++

	if idx < len(m.errors) && m.errors[idx] != nil {
		return nil, m.errors[idx]
	}

	if idx < len(m.responses) {
		return m.responses[idx], nil
	}

	return nil, fmt.Errorf("no more responses configured (call #%d)", idx)
}

func newTestClientWithSequentialMock(t *testing.T) (*Client, *sequentialMockHTTPClient) {
	cfg := Config{
		BaseURL:  "https://api.eva.team",
		APIToken: "test-token",
	}

	client, err := NewClient(&cfg)
	require.NoError(t, err)

	mockHTTP := &sequentialMockHTTPClient{}
	client.httpClient = mockHTTP

	return client, mockHTTP
}
