package warlot

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// authHeaders builds the authentication headers from the Client configuration.
func (c *Client) authHeaders() http.Header {
	h := http.Header{}
	if c.APIKey != "" {
		h.Set("x-api-key", c.APIKey)
	}
	if c.HolderID != "" {
		h.Set("x-holder-id", c.HolderID)
	}
	if c.ProjectName != "" {
		h.Set("x-project-name", c.ProjectName)
	}
	return h
}

// buildHeaders collects headers from CallOptions into a header map.
func buildHeaders(h http.Header, opts ...CallOption) http.Header {
	co := &callOptions{}
	for _, o := range opts {
		o(co)
	}
	if h == nil {
		h = http.Header{}
	}
	mergeHeaders(h, co.headers)
	return h
}

// mergeHeaders appends values from src into dst.
func mergeHeaders(dst http.Header, src http.Header) {
	if src == nil {
		return
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

// statusOf returns the HTTP status code or zero if the response is nil.
func statusOf(res *http.Response) int {
	if res == nil {
		return 0
	}
	return res.StatusCode
}

// parseAPIError decodes an error body and captures message/code/details when available.
func parseAPIError(code int, b []byte) *APIError {
	apiErr := &APIError{StatusCode: code, Body: string(b)}
	var msg struct {
		Message string      `json:"message"`
		Error   string      `json:"error"`
		Code    string      `json:"code"`
		Details interface{} `json:"details"`
	}
	if json.Unmarshal(b, &msg) == nil {
		if msg.Message != "" {
			apiErr.Message = msg.Message
		} else if msg.Error != "" {
			apiErr.Message = msg.Error
		}
		apiErr.Code = msg.Code
		apiErr.Details = msg.Details
	}
	return apiErr
}

// parseRetryAfter interprets Retry-After header values (seconds or HTTP-date).
func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

// redactHeaders masks sensitive header values for logging.
func redactHeaders(h http.Header) http.Header {
	if h == nil {
		return h
	}
	cp := http.Header{}
	for k, vs := range h {
		for _, v := range vs {
			if strings.EqualFold(k, "x-api-key") {
				if len(v) > 8 {
					cp.Add(k, v[:4]+"…"+v[len(v)-4:])
				} else {
					cp.Add(k, "********")
				}
			} else {
				cp.Add(k, v)
			}
		}
	}
	return cp
}

// randFloat64 returns a pseudo-random value in [0,1).
// It is based on the monotonic clock to avoid a global RNG.
func randFloat64() float64 {
	n := time.Now().UnixNano()
	n ^= n << 13
	n ^= n >> 7
	n ^= n << 17
	if n < 0 {
		n = -n
	}
	return float64(n%1000) / 1000.0
}

// normalizeBackoff ensures sane defaults for backoff windows.
func normalizeBackoff(initial, max time.Duration) (time.Duration, time.Duration) {
	if initial <= 0 {
		initial = 200 * time.Millisecond
	}
	if max <= 0 {
		max = 2 * time.Second
	}
	return initial, max
}

// normalizeRetries ensures non-negative retry counts.
func normalizeRetries(r int) int {
	if r < 0 {
		return 0
	}
	return r
}

// jitterSleep sleeps for a randomized duration based on the current backoff.
// Context cancellation is respected.
func jitterSleep(ctx context.Context, backoff, maxBack time.Duration) {
	jitter := time.Duration(float64(backoff) * (0.5 + 0.5*randFloat64()))
	if jitter > maxBack {
		jitter = maxBack
	}
	timer := time.NewTimer(jitter)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

// nextBackoff doubles backoff up to maxBack.
func nextBackoff(backoff, maxBack time.Duration) time.Duration {
	backoff *= 2
	if backoff > maxBack {
		backoff = maxBack
	}
	return backoff
}
