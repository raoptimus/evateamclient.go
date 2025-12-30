package evateamclient

import (
	"context"
	"fmt"
	"net/url"
	"runtime"
	"time"

	"github.com/imroc/req/v3"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

const defaultTimeout = 30 * time.Second
const basePath = "/api/"

var (
	ErrOptionIsRequired    = errors.New("option is required")
	ErrBodyIsRequired      = errors.New("body is required")
	ErrRPCMethodIsRequired = errors.New("RPCRequest.Method is required")
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Client структура
type Client struct {
	metrics    Metrics
	baseURL    *url.URL
	apiToken   string
	httpClient *req.Client
	logger     Logger
	debug      bool
}

// Config структура
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

	httpClient := req.C().
		SetTimeout(cfg.Timeout).
		SetCommonBearerAuthToken(cfg.APIToken).
		SetCommonHeader("Accept", "application/json").
		SetCommonHeader("Content-Type", "application/json")

	c := &Client{
		baseURL:    baseURL.JoinPath(basePath),
		apiToken:   cfg.APIToken,
		httpClient: httpClient,
		debug:      cfg.Debug,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
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
			fmt.Println("RESPONSE:\n")
			fmt.Println(string(respBodyBytes))
		}
	}()

	reqBodyBytes, err = json.Marshal(body)
	if err != nil {
		return errors.WithMessage(err, "marshal request body")
	}

	request := c.httpClient.R().
		SetContext(ctx).
		SetBodyBytes(reqBodyBytes)

	resp, err := request.Post(reqURL)
	if err != nil {
		return errors.WithMessage(err, "http request failed")
	}

	statusCode = resp.StatusCode
	respBodyBytes = resp.Bytes()

	if resp.IsErrorState() {
		return errors.Errorf("API error %d: %s", resp.StatusCode, string(resp.Bytes()))
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

func (c *Client) logInfo(ctx context.Context, msg string, args ...any) {
	if c.logger != nil {
		c.logger.Info(ctx, msg, args...)
	}
}

func (c *Client) logWarn(ctx context.Context, msg string, args ...any) {
	if c.logger != nil {
		c.logger.Warn(ctx, msg, args...)
	}
}

func (c *Client) logError(ctx context.Context, msg string, args ...any) {
	if c.logger != nil {
		c.logger.Error(ctx, msg, args...)
	}
}

// Close закрывает клиент
func (c *Client) Close() error {
	return nil
}

// functionName вспомогательная функция (БЕЗ Get префикса!)
func functionName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	return runtime.FuncForPC(pc).Name()
}
