package dffunc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	client = &http.Client{
		Transport: &http.Transport{
			DialContext: ((&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext),
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: time.Minute * 5,
	}
)

const ApiTmpl = "/api/v1/func/guance_siem__api.%s"

type DFCall interface {
	Call(name string, kwargs map[string]any) map[string]any
}

var _ DFCall = (*DFFunc)(nil)

type DFFunc struct {
	BaseURL string
}

func NewDFFunc(funcURL string) *DFFunc {
	return &DFFunc{
		BaseURL: funcURL,
	}
}

func (f *DFFunc) Call(name string, kwargs map[string]any) map[string]any {
	r := map[string]any{}
	urlPath, err := url.JoinPath(f.BaseURL, fmt.Sprintf(ApiTmpl, name))
	if err != nil {
		r["error"] = err.Error()
		return r
	}

	queryBody := map[string]any{
		"kwargs": kwargs,
	}
	b, err := json.Marshal(queryBody)
	if err != nil {
		r["error"] = err.Error()
		return r
	}

	req, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(b))
	if err != nil {
		r["error"] = err.Error()
		return r
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		r["error"] = err.Error()
		return r
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		r["error"] = err.Error()
		return r
	}

	if resp.StatusCode/100 != 2 {
		r["ok"] = false
		r["error"] = fmt.Sprintf("call function failed, api status code %d", resp.StatusCode)

		var body Body
		if err := json.Unmarshal(buf, &body); err == nil {
			r["message"] = body.Message
			r["reason"] = body.Reason
			r["ok"] = body.Ok
			r["detail"] = map[string]any{
				"exeception": body.Detail.Exeception,
			}
		}
		return r
	}

	var body Body

	if err := json.Unmarshal(buf, &body); err != nil {
		r["error"] = err.Error()
		return r
	}

	r["ok"] = true
	r["message"] = ""
	r["reason"] = ""
	r["error"] = ""
	if v, ok := body.Data["result"]; ok {
		r["result"] = v
	}

	return r
}

type Body struct {
	Ok     bool    `json:"ok"`
	Error  float64 `json:"error"`
	Reason string  `json:"reason"`
	Detail struct {
		Exeception string `json:"exception"`
		Traceback  string `json:"traceback"`
	} `json:"detail"`

	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}
