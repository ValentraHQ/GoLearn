package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Remote talks to a runner daemon (cmd/runner-daemon) over HTTP. It lets the
// internet-facing API run with no container-runtime access at all: only the
// daemon, on an isolated network, can start sandbox containers.
type Remote struct {
	base   string
	token  string
	client *http.Client

	mu   sync.Mutex
	last Status
	at   time.Time
}

func NewRemote(baseURL, token string) *Remote {
	return &Remote{
		base:   strings.TrimRight(baseURL, "/"),
		token:  token,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (r *Remote) do(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, r.base+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+r.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return r.client.Do(req)
}

func (r *Remote) Status(ctx context.Context) Status {
	r.mu.Lock()
	defer r.mu.Unlock()
	if time.Since(r.at) < 10*time.Second {
		return r.last
	}
	st := Status{Backend: "remote"}
	cctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	resp, err := r.do(cctx, http.MethodGet, "/v1/status", nil)
	switch {
	case err != nil:
		st.Reason = "The code runner service is not reachable."
	case resp.StatusCode != http.StatusOK:
		resp.Body.Close()
		st.Reason = fmt.Sprintf("The code runner service returned %s.", resp.Status)
	default:
		defer resp.Body.Close()
		var remote Status
		if json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&remote) == nil && remote.Available {
			st.Available = true
		} else {
			st.Reason = remote.Reason
			if st.Reason == "" {
				st.Reason = "The code runner service reports it is unavailable."
			}
		}
	}
	r.last, r.at = st, time.Now()
	return st
}

func (r *Remote) Run(ctx context.Context, req Request) (Result, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return Result{}, err
	}
	resp, err := r.do(ctx, http.MethodPost, "/v1/run", body)
	if err != nil {
		return Result{}, fmt.Errorf("%w: runner service unreachable", ErrUnavailable)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusServiceUnavailable:
		return Result{}, ErrBusy
	case http.StatusUnauthorized, http.StatusForbidden:
		return Result{}, fmt.Errorf("%w: runner rejected credentials", ErrUnavailable)
	default:
		return Result{}, errors.New("runner service error: " + resp.Status)
	}
	var res Result
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&res); err != nil {
		return Result{}, fmt.Errorf("runner returned malformed output: %w", err)
	}
	return res, nil
}
