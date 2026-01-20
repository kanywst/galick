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
func NewScriptAttacker(scriptPath string, timeout time.Duration, insecure bool) (protocols.Attacker, error) {
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
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		MaxIdleConns:    1000,
		MaxConnsPerHost: 1000,
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

// Attack performs a single request by executing the Starlark script.
func (s *Attacker) Attack(ctx context.Context) metrics.Result {
	start := time.Now()

	thread := &starlark.Thread{Name: "worker"}
	
	res, err := starlark.Call(thread, s.requestFn, nil, nil)
	if err != nil {
		return metrics.Result{Timestamp: start, Error: fmt.Sprintf("script error: %v", err), Latency: time.Since(start)}
	}

	dict, ok := res.(*starlark.Dict)
	if !ok {
		return metrics.Result{Timestamp: start, Error: "script must return a dict", Latency: time.Since(start)}
	}

	method := "GET"
	url := ""
	var body io.Reader

	for _, item := range dict.Items() {
		k, ok := item[0].(starlark.String)
		if !ok {
			continue
		}
		switch k.String() {
		case "method":
			if v, ok := item[1].(starlark.String); ok {
				method = string(v)
			}
		case "url":
			if v, ok := item[1].(starlark.String); ok {
				url = string(v)
			}
		case "body":
			if v, ok := item[1].(starlark.String); ok {
				body = strings.NewReader(string(v))
			}
		}
	}

	if url == "" {
		return metrics.Result{Timestamp: start, Error: "script returned empty url", Latency: time.Since(start)}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return metrics.Result{Timestamp: start, Error: err.Error(), Latency: time.Since(start)}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return metrics.Result{Timestamp: start, Error: err.Error(), Latency: time.Since(start)}
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