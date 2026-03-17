// Package loadhttp implements the HTTP load testing protocol.
package loadhttp

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"time"

	"github.com/kanywst/galick/pkg/metrics"
	"github.com/kanywst/galick/pkg/protocols"
)

// Attacker implements the Attacker interface for HTTP.
type Attacker struct {
	client  *http.Client
	method  string
	url     string
	headers map[string]string
}

// NewAttacker creates a new HTTP Attacker.
func NewAttacker(method, url string, timeout time.Duration, insecure bool, headers map[string]string, workers int) protocols.Attacker {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: insecure},
		MaxIdleConns:       workers,
		MaxIdleConnsPerHost: workers,
	}

	copiedHeaders := make(map[string]string, len(headers))
	for k, v := range headers {
		copiedHeaders[k] = v
	}

	return &Attacker{
		client: &http.Client{
			Transport: tr,
			Timeout:   timeout,
		},
		method:  method,
		url:     url,
		headers: copiedHeaders,
	}
}

// Name returns the name of the protocol.
func (h *Attacker) Name() string {
	return "http"
}

// Attack performs a single HTTP request.
func (h *Attacker) Attack(ctx context.Context) metrics.Result {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, h.method, h.url, nil)
	if err != nil {
		return metrics.Result{
			Timestamp: start,
			Error:     err.Error(),
			Latency:   time.Since(start),
		}
	}

	for k, v := range h.headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return metrics.Result{
			Timestamp: start,
			Error:     err.Error(),
			Latency:   time.Since(start),
		}
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