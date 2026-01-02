package evateamclient

import (
	"context"
	"net/url"
	"os"
	"time"

	"github.com/imroc/req/v3"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

const defaultTimeout = 30 * time.Second
const basePath = "/api/"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type HTTPClient interface {
	Post(ctx context.Context, body []byte, url string) (*req.Response, error)
}
type httpClient struct {
	hc *req.Client
}

func (h *httpClient) Post(ctx context.Context, body []byte, url string) (*req.Response, error) {
	return h.hc.R().
		SetContext(ctx).
		SetBodyBytes(body).
		Post(url)
}

// Client is the EVA Team API client
type Client struct {
	metrics    Metrics
	baseURL    *url.URL
	apiToken   string
	httpClient HTTPClient
	logger     Logger
	debug      bool
}

// Config holds client configuration
type Config struct {
	BaseURL  string
	APIToken string
	Debug    bool
	Timeout  time.Duration
}

type Option func(*Client)

func WithLogger(l Logger) Option {
	return func(c *Client) {
		c.logger = l
	}
}

func WithDebug(debug bool) Option {
	return func(c *Client) {
		c.debug = debug
	}
}

func WithMetrics(m Metrics) Option {
	return func(c *Client) {
		c.metrics = m
	}
}

func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, errors.WithMessage(errors.WithStack(ErrOptionIsRequired), "baseURL")
	}
	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, errors.WithMessage(err, "baseURL")
	}

	if cfg.APIToken == "" {
		return nil, errors.WithMessage(errors.WithStack(ErrOptionIsRequired), "APIToken")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}

	hc := req.C().
		SetTimeout(cfg.Timeout).
		SetCommonBearerAuthToken(cfg.APIToken).
		SetCommonHeader("Accept", "application/json").
		SetCommonHeader("Content-Type", "application/json")

	c := &Client{
		baseURL:    baseURL.JoinPath(basePath),
		apiToken:   cfg.APIToken,
		httpClient: &httpClient{hc: hc},
		debug:      cfg.Debug,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Close closes client
func (c *Client) Close() error {
	return nil
}

func (c *Client) doRequest(ctx context.Context, body *RPCRequest, result any) error {
	if body == nil {
		return errors.WithStack(ErrBodyIsRequired)
	}
	if body.Method == "" {
		return errors.WithStack(ErrRPCMethodIsRequired)
	}

	reqURL := c.baseURL.String() + "?m=" + url.QueryEscape(body.Method)
	fname := functionName(2)
	startTime := time.Now()

	var (
		err           error
		reqBodyBytes  []byte
		respBodyBytes []byte
		statusCode    int
	)

	defer func() {
		if c.metrics != nil {
			c.metrics.RecordRequestDuration(statusCode, body.Method, c.baseURL.Host, fname, time.Since(startTime).Seconds())
		}
		if c.debug {
			c.logDebug(ctx, "Request",
				"method", body.Method,
				"url", reqURL,
				"func", fname,
				"requestBody", string(reqBodyBytes),
				"responseBody", string(respBodyBytes),
				"responseStatus", statusCode,
				"duration", time.Since(startTime).String(),
				"error", err,
			)

			f, err := os.OpenFile("./response.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				c.logError(ctx, err.Error())
				return
			}
			defer f.Close()
			if _, err = f.Write(respBodyBytes); err != nil {
				c.logError(ctx, err.Error())
			}
		}
	}()

	reqBodyBytes, err = json.Marshal(body)
	if err != nil {
		return errors.WithMessage(err, "marshal request body")
	}

	resp, err := c.httpClient.Post(ctx, reqBodyBytes, reqURL)
	if err != nil {
		return errors.WithMessage(err, "http request failed")
	}

	statusCode = resp.StatusCode
	respBodyBytes = resp.Bytes()

	if resp.IsErrorState() {
		return errors.Errorf("API error %d: %s", resp.StatusCode, string(resp.Bytes()))
	}

	// Check for RPC error in 200 OK response
	var rpcErr rpcErrorResponse
	if err := json.Unmarshal(respBodyBytes, &rpcErr); err != nil {
		return errors.WithMessage(err, "unmarshal rpc error check")
	}
	if rpcErr.Error != nil {
		return errors.WithMessagef(rpcErr.Error, "RPC error %d", rpcErr.Error.Code)
	}

	if result != nil {
		if err := json.Unmarshal(respBodyBytes, result); err != nil {
			return errors.WithMessage(err, "unmarshal response body")
		}
	}

	return nil
}

func (c *Client) logDebug(ctx context.Context, msg string, args ...any) {
	if c.logger != nil && c.debug {
		c.logger.Debug(ctx, msg, args...)
	}
}

func (c *Client) logError(ctx context.Context, msg string, args ...any) {
	if c.logger != nil {
		c.logger.Error(ctx, msg, args...)
	}
}
