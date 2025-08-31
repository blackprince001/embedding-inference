package wrapper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"teiwrappergolang/internal/config"
	"teiwrappergolang/internal/domain/entities"
	"teiwrappergolang/internal/domain/errors"
	"teiwrappergolang/internal/infrastructure/logging"

	"go.uber.org/zap"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	timeout    time.Duration
	maxRetries int
	retryDelay time.Duration
	logger     *logging.Logger
	userAgent  string
}

func NewHTTPClient(cfg *config.TEIConfig, logger *logging.Logger) (*Client, error) {
	parsedURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        cfg.MaxConnections,
		MaxIdleConnsPerHost: cfg.MaxConnections,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   false,
		DisableCompression:  false,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimSuffix(parsedURL.String(), "/"),
		timeout:    cfg.Timeout,
		maxRetries: cfg.MaxRetries,
		retryDelay: cfg.RetryDelay,
		logger:     logger,
	}, nil
}

func (c *Client) Get(ctx context.Context, endpoint string) ([]byte, error) {
	url := c.baseURL + endpoint

	c.logger.Debug("GET request",
		zap.String("url", url),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setDefaultHeaders(req)

	return c.executeWithRetry(ctx, req)
}

func (c *Client) Post(ctx context.Context, endpoint string, body any) ([]byte, error) {
	url := c.baseURL + endpoint

	c.logger.Debug("POST request",
		zap.String("url", url),
		zap.String("body_type", fmt.Sprintf("%T", body)),
	)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setDefaultHeaders(req)
	req.Header.Set(entities.HeaderContentType, entities.ContentTypeJSON)

	return c.executeWithRetry(ctx, req)
}

func (c *Client) PostRaw(ctx context.Context, endpoint string, body []byte, contentType string) ([]byte, error) {
	url := c.baseURL + endpoint

	c.logger.Debug("POST raw request",
		zap.String("url", url),
		zap.String("content_type", contentType),
		zap.Int("body_size", len(body)),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setDefaultHeaders(req)
	req.Header.Set(entities.HeaderContentType, contentType)

	return c.executeWithRetry(ctx, req)
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.httpClient.Timeout = timeout
	c.logger.Debug("HTTP client timeout updated", zap.Duration("timeout", timeout))
}

func (c *Client) Close() error {
	c.logger.Debug("Closing HTTP client")
	c.httpClient.CloseIdleConnections()

	return nil
}

func (c *Client) setDefaultHeaders(req *http.Request) {
	req.Header.Set(entities.HeaderUserAgent, c.userAgent)
	req.Header.Set(entities.HeaderAccept, entities.ContentTypeJSON)
}

func (c *Client) executeWithRetry(ctx context.Context, req *http.Request) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.calculateRetryDelay(attempt)):
			}

			c.logger.Debug("Retrying request",
				zap.Int("attempt", attempt),
				zap.String("url", req.URL.String()),
			)
		}

		if req.Body != nil {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read request body: %w", err)
			}
			req.Body = io.NopCloser(bytes.NewReader(body))
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = c.wrapNetworkError(err)

			if teiErr, ok := lastErr.(*errors.TEIError); ok && teiErr.IsRetryable() {
				c.logger.Warn("Request failed, will retry",
					zap.Error(err),
					zap.Int("attempt", attempt),
				)
				continue
			}

			return nil, lastErr
		}

		responseBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			c.logger.Debug("Request completed successfully",
				zap.String("url", req.URL.String()),
				zap.Int("status_code", resp.StatusCode),
				zap.Int("response_size", len(responseBody)),
				zap.Int("attempt", attempt+1),
			)
			return responseBody, nil
		}
		lastErr = c.handleErrorResponse(resp.StatusCode, responseBody)

		if teiErr, ok := lastErr.(*errors.TEIError); ok && teiErr.IsRetryable() {
			c.logger.Warn("Request failed with retryable error",
				zap.Error(lastErr),
				zap.Int("status_code", resp.StatusCode),
				zap.Int("attempt", attempt),
			)
			continue
		}

		return nil, lastErr
	}

	c.logger.Error("Request failed after all retries",
		zap.Error(lastErr),
		zap.String("url", req.URL.String()),
		zap.Int("max_retries", c.maxRetries),
	)

	return nil, fmt.Errorf("request failed after %d retries: %w", c.maxRetries, lastErr)
}

func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	c.logger.Debug("Handling error response",
		zap.Int("status_code", statusCode),
		zap.Int("body_size", len(body)),
	)

	var teiError struct {
		Error     string `json:"error"`
		ErrorType string `json:"error_type"`
		Message   string `json:"message"`
		Type      string `json:"type"`
	}

	message := fmt.Sprintf("HTTP %d", statusCode)

	if len(body) > 0 {
		if err := json.Unmarshal(body, &teiError); err == nil {
			if teiError.Error != "" {
				message = teiError.Error
			} else if teiError.Message != "" {
				message = teiError.Message
			}
		} else {
			message = string(body)
			if len(message) > 200 {
				message = message[:200] + "..."
			}
		}
	}

	return errors.NewTEIErrorFromHTTP(statusCode, message)
}

func (c *Client) wrapNetworkError(err error) error {
	if err == nil {
		return nil
	}

	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			return errors.NewTEIError(err.Error(), errors.ErrorTypeTimeout)
		}
	}

	if err == context.Canceled {
		return errors.NewTEIError("request canceled", errors.ErrorTypeTimeout)
	}

	if err == context.DeadlineExceeded {
		return errors.NewTEIError("request timeout", errors.ErrorTypeTimeout)
	}

	if strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "no such host") ||
		strings.Contains(err.Error(), "network is unreachable") {
		return errors.NewTEIError(err.Error(), errors.ErrorTypeNetwork)
	}

	return errors.NewTEIError(err.Error(), errors.ErrorTypeNetwork)
}

func (c *Client) calculateRetryDelay(attempt int) time.Duration {
	baseDelay := c.retryDelay
	exponentialDelay := time.Duration(1<<uint(attempt-1)) * baseDelay

	maxDelay := 30 * time.Second
	if exponentialDelay > maxDelay {
		exponentialDelay = maxDelay
	}

	return exponentialDelay
}
