// Copyright 2026 Kava Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package clientid forwards a client ID header from incoming Rosetta requests
// onto the node RPC calls made while serving them.
//
// Middleware stores the incoming request's value in the request context, and
// Transport adds it to each outbound node call made with that context. The value
// is added per call and never set on a shared client or connection, so one
// caller's ID cannot appear on another caller's node calls.
package clientid

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/http/httpguts"
)

// refusedHeaders are header names that would break or change the node call.
var refusedHeaders = map[string]struct{}{
	"accept":            {},
	"connection":        {},
	"content-length":    {},
	"content-type":      {},
	"host":              {},
	"transfer-encoding": {},
	"upgrade":           {},
}

// Config configures client ID forwarding. The zero value turns it off.
type Config struct {
	// Header is the lowercase header name, for example "x-client-id".
	Header string
	// Default is sent on node calls made outside an incoming request.
	Default string
}

// Enabled reports whether a header name is configured.
func (c Config) Enabled() bool {
	return c.Header != ""
}

// ParseConfig validates and normalizes a header name and default value.
// An empty header turns the feature off; a default then is an error.
func ParseConfig(header, defaultValue string) (Config, error) {
	header = strings.ToLower(strings.TrimSpace(header))
	defaultValue = strings.TrimSpace(defaultValue)

	if header == "" {
		if defaultValue != "" {
			return Config{}, fmt.Errorf("client ID default is set without a header name")
		}
		return Config{}, nil
	}

	if !httpguts.ValidHeaderFieldName(header) {
		return Config{}, fmt.Errorf("invalid client ID header name %q", header)
	}

	if _, refused := refusedHeaders[header]; refused || strings.HasPrefix(header, "sec-websocket-") {
		return Config{}, fmt.Errorf("client ID header name %q is not allowed", header)
	}

	// The value is not included in the error so it is never logged.
	if !httpguts.ValidHeaderFieldValue(defaultValue) {
		return Config{}, fmt.Errorf("invalid client ID default value")
	}

	return Config{Header: header, Default: defaultValue}, nil
}

type contextKey struct{}

// requestValue marks a context as belonging to an incoming request. An empty
// value means the request had no client ID header.
type requestValue struct {
	value string
}

// WithRequestValue returns a context carrying an incoming request's client ID.
// Pass an empty value for a request without the header.
func WithRequestValue(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, contextKey{}, requestValue{value: value})
}

func fromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(contextKey{}).(requestValue)
	return v.value, ok
}

// Middleware stores each request's client ID header in the request context.
// It returns next unchanged when cfg is not enabled.
func Middleware(cfg Config, next http.Handler) http.Handler {
	if !cfg.Enabled() {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := strings.TrimSpace(r.Header.Get(cfg.Header))
		if !httpguts.ValidHeaderFieldValue(value) {
			value = ""
		}

		next.ServeHTTP(w, r.WithContext(WithRequestValue(r.Context(), value)))
	})
}

// Transport wraps base to add the client ID header to outbound requests:
// the incoming request's value inside a request, nothing for a request
// without the header, and cfg.Default outside a request. A header already
// set on the outbound request is kept. It returns base unchanged when cfg
// is not enabled, and uses http.DefaultTransport when base is nil.
func Transport(cfg Config, base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	if !cfg.Enabled() {
		return base
	}

	return &transport{cfg: cfg, base: base}
}

type transport struct {
	cfg  Config
	base http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get(t.cfg.Header) != "" {
		return t.base.RoundTrip(req)
	}

	value, inRequest := fromContext(req.Context())
	if !inRequest {
		value = t.cfg.Default
	}

	if value == "" {
		return t.base.RoundTrip(req)
	}

	// A RoundTripper must not modify the request it is given.
	out := req.Clone(req.Context())
	out.Header.Set(t.cfg.Header, value)

	return t.base.RoundTrip(out)
}
