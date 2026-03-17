// Package script implements the Starlark scripting protocol for load testing.
package script

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kanywst/galick/pkg/metrics"
	"github.com/kanywst/galick/pkg/protocols"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// Attacker runs a Starlark script to generate requests.
type Attacker struct {
	client    *http.Client
	requestFn starlark.Value
}

// NewScriptAttacker creates a new Starlark script attacker.
func NewScriptAttacker(scriptPath string, timeout time.Duration, insecure bool, workers int) (protocols.Attacker, error) {
	thread := &starlark.Thread{Name: "main"}
	opts := &syntax.FileOptions{}
	globals, err := starlark.ExecFileOptions(opts, thread, scriptPath, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("script execution failed: %w", err)
	}

	reqFn, ok := globals["request"]
	if !ok {
		return nil, fmt.Errorf("script must define a 'request()' function")
	}

	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: insecure},
		MaxIdleConns:       workers,
		MaxIdleConnsPerHost: workers,
	}

	return &Attacker{
		client: &http.Client{
			Transport: tr,
			Timeout:   timeout,
		},
		requestFn: reqFn,
	}, nil
}

// Name returns the name of the protocol.
func (s *Attacker) Name() string {
	return "script"
}

// requestSpec holds the parsed fields from a Starlark request() return value.
type requestSpec struct {
	method  string
	url     string
	body    io.Reader
	headers *starlark.Dict
}

// parseRequestDict extracts method, url, body, and headers from a Starlark dict.
func parseRequestDict(dict *starlark.Dict) (requestSpec, error) {
	spec := requestSpec{method: "GET"}
	for _, item := range dict.Items() {
		k, ok := item[0].(starlark.String)
		if !ok {
			continue
		}
		switch k.GoString() {
		case "method":
			v, ok := item[1].(starlark.String)
			if !ok {
				return spec, fmt.Errorf("script returned 'method' with non-string value")
			}
			spec.method = v.GoString()
		case "url":
			v, ok := item[1].(starlark.String)
			if !ok {
				return spec, fmt.Errorf("script returned 'url' with non-string value")
			}
			spec.url = v.GoString()
		case "body":
			v, ok := item[1].(starlark.String)
			if !ok {
				return spec, fmt.Errorf("script returned 'body' with non-string value")
			}
			spec.body = strings.NewReader(v.GoString())
		case "headers":
			v, ok := item[1].(*starlark.Dict)
			if !ok {
				return spec, fmt.Errorf("script returned 'headers' with non-dict value")
			}
			spec.headers = v
		}
	}
	return spec, nil
}

// applyHeaders sets headers from a Starlark dict onto an http.Request.
func applyHeaders(req *http.Request, headers *starlark.Dict) error {
	for _, item := range headers.Items() {
		k, ok := item[0].(starlark.String)
		if !ok {
			return fmt.Errorf("script returned header with non-string key")
		}
		v, ok := item[1].(starlark.String)
		if !ok {
			return fmt.Errorf("script returned header with non-string value")
		}
		keyStr := k.GoString()
		if keyStr == "" {
			return fmt.Errorf("script returned header with empty key")
		}
		req.Header.Set(keyStr, v.GoString())
	}
	return nil
}

// errResult creates an error metrics.Result from the given start time and message.
func errResult(start time.Time, msg string) metrics.Result {
	return metrics.Result{Timestamp: start, Error: msg, Latency: time.Since(start)}
}

// Attack performs a single request by executing the Starlark script.
func (s *Attacker) Attack(ctx context.Context) metrics.Result {
	start := time.Now()

	thread := &starlark.Thread{Name: "worker"}

	res, err := starlark.Call(thread, s.requestFn, nil, nil)
	if err != nil {
		return errResult(start, fmt.Sprintf("script error: %v", err))
	}

	dict, ok := res.(*starlark.Dict)
	if !ok {
		return errResult(start, "script must return a dict")
	}

	spec, err := parseRequestDict(dict)
	if err != nil {
		return errResult(start, err.Error())
	}
	if spec.url == "" {
		return errResult(start, "script returned empty url")
	}

	req, err := http.NewRequestWithContext(ctx, spec.method, spec.url, spec.body)
	if err != nil {
		return errResult(start, err.Error())
	}

	if spec.headers != nil {
		if err := applyHeaders(req, spec.headers); err != nil {
			return errResult(start, err.Error())
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return errResult(start, err.Error())
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read body to count bytes. io.Copy to io.Discard is efficient and streams
	// the data without loading the entire body into memory.
	written, _ := io.Copy(io.Discard, resp.Body)

	return metrics.Result{
		Timestamp: start,
		Latency:   time.Since(start),
		BytesIn:   uint64(written),
		Code:      uint16(resp.StatusCode),
	}
}