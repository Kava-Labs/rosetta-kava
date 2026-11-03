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

package clientid_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/kava-labs/rosetta-kava/clientid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const header = "x-client-id"

func TestParseConfig(t *testing.T) {
	tests := map[string]struct {
		header, def string
		want        clientid.Config
		wantErr     string
	}{
		"unset is off":                {want: clientid.Config{}},
		"header only":                 {header: header, want: clientid.Config{Header: header}},
		"header and default":          {header: header, def: "svc", want: clientid.Config{Header: header, Default: "svc"}},
		"lowercase and trim":          {header: "  X-Client-ID ", def: " svc ", want: clientid.Config{Header: header, Default: "svc"}},
		"default without header":      {def: "svc", wantErr: "without a header name"},
		"invalid name":                {header: "x client id", wantErr: "invalid client ID header name"},
		"refused name":                {header: "Host", wantErr: "not allowed"},
		"refused websocket name":      {header: "sec-websocket-key", wantErr: "not allowed"},
		"invalid default":             {header: header, def: "a\nb", wantErr: "invalid client ID default value"},
		"invalid default not printed": {header: header, def: "secret\x00", wantErr: "invalid client ID default value"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := clientid.ParseConfig(tc.header, tc.def)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.NotContains(t, err.Error(), "secret")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.want.Header != "", got.Enabled())
		})
	}
}

// fakeNode records the client ID header and any other headers it receives.
type fakeNode struct {
	mu       sync.Mutex
	received []http.Header
}

func (n *fakeNode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.mu.Lock()
	n.received = append(n.received, r.Header.Clone())
	n.mu.Unlock()
	// Echo the header so each caller can check its own call.
	_, _ = io.WriteString(w, r.Header.Get(header))
}

func (n *fakeNode) last() http.Header {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.received[len(n.received)-1]
}

// setup returns a Rosetta-like server whose handler makes one node call with
// the request context and writes the node's response.
func setup(t *testing.T, cfg clientid.Config) (*fakeNode, *httptest.Server) {
	node := &fakeNode{}
	nodeServer := httptest.NewServer(node)
	t.Cleanup(nodeServer.Close)

	nodeClient := &http.Client{Transport: clientid.Transport(cfg, nil)}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, nodeServer.URL, strings.NewReader("{}"))
		require.NoError(t, err)
		res, err := nodeClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = res.Body.Close() }()
		_, _ = io.Copy(w, res.Body)
	})

	rosetta := httptest.NewServer(clientid.Middleware(cfg, handler))
	t.Cleanup(rosetta.Close)

	return node, rosetta
}

func call(t *testing.T, url string, headers map[string]string) string {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader("{}"))
	require.NoError(t, err)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return string(body)
}

func TestForwardsRequestValue(t *testing.T) {
	node, rosetta := setup(t, clientid.Config{Header: header, Default: "svc"})

	got := call(t, rosetta.URL, map[string]string{header: "caller-a", "x-other": "dropped"})

	assert.Equal(t, "caller-a", got)
	assert.Empty(t, node.last().Get("x-other"), "only the client ID header is forwarded")
}

func TestRequestWithoutHeaderSendsNone(t *testing.T) {
	node, rosetta := setup(t, clientid.Config{Header: header, Default: "svc"})

	call(t, rosetta.URL, nil)

	_, present := node.last()[http.CanonicalHeaderKey(header)]
	assert.False(t, present, "the default is not used for a request without the header")
}

func TestDefaultOutsideRequest(t *testing.T) {
	node := &fakeNode{}
	nodeServer := httptest.NewServer(node)
	defer nodeServer.Close()

	client := &http.Client{Transport: clientid.Transport(clientid.Config{Header: header, Default: "svc"}, nil)}
	res, err := client.Post(nodeServer.URL, "application/json", strings.NewReader("{}"))
	require.NoError(t, err)
	_ = res.Body.Close()

	assert.Equal(t, "svc", node.last().Get(header))
}

func TestNoDefaultOutsideRequestSendsNone(t *testing.T) {
	node := &fakeNode{}
	nodeServer := httptest.NewServer(node)
	defer nodeServer.Close()

	client := &http.Client{Transport: clientid.Transport(clientid.Config{Header: header}, nil)}
	res, err := client.Post(nodeServer.URL, "application/json", strings.NewReader("{}"))
	require.NoError(t, err)
	_ = res.Body.Close()

	_, present := node.last()[http.CanonicalHeaderKey(header)]
	assert.False(t, present)
}

func TestDisabledChangesNothing(t *testing.T) {
	base := http.DefaultTransport
	assert.Same(t, base, clientid.Transport(clientid.Config{}, base))

	node, rosetta := setup(t, clientid.Config{})
	call(t, rosetta.URL, map[string]string{header: "caller-a"})

	_, present := node.last()[http.CanonicalHeaderKey(header)]
	assert.False(t, present)
}

func TestExplicitHeaderWins(t *testing.T) {
	node := &fakeNode{}
	nodeServer := httptest.NewServer(node)
	defer nodeServer.Close()

	client := &http.Client{Transport: clientid.Transport(clientid.Config{Header: header, Default: "svc"}, nil)}
	ctx := clientid.WithRequestValue(context.Background(), "caller-a")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, nodeServer.URL, strings.NewReader("{}"))
	require.NoError(t, err)
	req.Header.Set(header, "explicit")
	res, err := client.Do(req)
	require.NoError(t, err)
	_ = res.Body.Close()

	assert.Equal(t, "explicit", node.last().Get(header))
}

func TestDoesNotModifyOutboundRequest(t *testing.T) {
	node := &fakeNode{}
	nodeServer := httptest.NewServer(node)
	defer nodeServer.Close()

	client := &http.Client{Transport: clientid.Transport(clientid.Config{Header: header}, nil)}
	ctx := clientid.WithRequestValue(context.Background(), "caller-a")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, nodeServer.URL, strings.NewReader("{}"))
	require.NoError(t, err)
	res, err := client.Do(req)
	require.NoError(t, err)
	_ = res.Body.Close()

	assert.Equal(t, "caller-a", node.last().Get(header))
	assert.Empty(t, req.Header.Get(header))
}

func TestConcurrentRequestsKeepTheirOwnValue(t *testing.T) {
	_, rosetta := setup(t, clientid.Config{Header: header, Default: "svc"})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			headers := map[string]string{}
			want := ""
			if i%3 != 0 {
				want = fmt.Sprintf("caller-%d", i)
				headers[header] = want
			}
			assert.Equal(t, want, call(t, rosetta.URL, headers))
		}(i)
	}
	wg.Wait()
}
