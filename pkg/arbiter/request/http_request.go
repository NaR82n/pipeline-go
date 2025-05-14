package request

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var defaultTransport = &http.Transport{
	DialContext: ((&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext),

	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func SetTransport(dialer *FilteringDialer) {
	tp := &http.Transport{
		DialContext: ((&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext),

		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if dialer != nil {
		tp.DialContext = dialer.DialContext
	}
	defaultTransport = tp
}

func NewHTTPClient(timeout time.Duration) *HTTPClient {
	if timeout == 0 {
		timeout = time.Minute * 2
	}

	return &HTTPClient{
		timeout: timeout,
	}
}

type HTTPCli interface {
	Request(method, url string, headers map[string]any, body any) (map[string]any, error)
}

type HTTPClient struct {
	timeout time.Duration
}

func (c *HTTPClient) Request(method, url string, headers map[string]any, body any) (map[string]any, error) {
	req, err := http.NewRequest(method, url, buildBody(body))
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		if v, ok := v.(string); ok {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{
		Transport: defaultTransport,
		Timeout:   c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	hdrs := map[string]any{}
	for k, v := range resp.Header {
		hdrs[k] = strings.Join(v, ", ")
	}

	return map[string]any{
		"status_code": resp.StatusCode,
		"body":        string(respBody),
		"headers":     hdrs,
	}, nil
}

func buildBody(val any) io.Reader {
	switch val := val.(type) {
	case string:
		return strings.NewReader(val)
	case []any:
		if val, err := json.Marshal(val); err == nil {
			return bytes.NewReader(val)
		}
	case map[string]any:
		if val, err := json.Marshal(val); err == nil {
			return bytes.NewReader(val)
		}
	case float64:
		return strings.NewReader(strconv.FormatFloat(val, 'f', -1, 64))
	case int64:
		return strings.NewReader(strconv.FormatInt(val, 10))
	case bool:
		return strings.NewReader(strconv.FormatBool(val))
	default:
	}
	return nil
}
