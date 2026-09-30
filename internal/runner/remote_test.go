package runner

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeDaemon(t *testing.T, token string, status int, res Result) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer "+token {
				http.Error(w, "no", http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}
	mux.HandleFunc("GET /v1/status", auth(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Status{Available: true, Backend: "docker"})
	}))
	mux.HandleFunc("POST /v1/run", auth(func(w http.ResponseWriter, r *http.Request) {
		if status != http.StatusOK {
			http.Error(w, "x", status)
			return
		}
		var req Request
		_ = json.NewDecoder(r.Body).Decode(&req)
		res.Stdout = req.Files["main.go"]
		_ = json.NewEncoder(w).Encode(res)
	}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestRemoteRun(t *testing.T) {
	srv := fakeDaemon(t, "secret-token-123456", 200, Result{ExitCode: 0})
	r := NewRemote(srv.URL, "secret-token-123456")
	if st := r.Status(context.Background()); !st.Available || st.Backend != "remote" {
		t.Fatalf("status = %+v", st)
	}
	res, err := r.Run(context.Background(), Request{Kind: KindRun, Files: map[string]string{"main.go": "hello"}})
	if err != nil || res.Stdout != "hello" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestRemoteBadTokenIsUnavailable(t *testing.T) {
	srv := fakeDaemon(t, "right-token-1234567", 200, Result{})
	r := NewRemote(srv.URL, "wrong-token-1234567")
	if st := r.Status(context.Background()); st.Available {
		t.Error("wrong token must not report available")
	}
	if _, err := r.Run(context.Background(), Request{Kind: KindRun}); !errors.Is(err, ErrUnavailable) {
		t.Errorf("want ErrUnavailable, got %v", err)
	}
}

func TestRemoteBusyAndDown(t *testing.T) {
	srv := fakeDaemon(t, "tok-tok-tok-tok-1", http.StatusServiceUnavailable, Result{})
	r := NewRemote(srv.URL, "tok-tok-tok-tok-1")
	if _, err := r.Run(context.Background(), Request{Kind: KindRun}); !errors.Is(err, ErrBusy) {
		t.Errorf("want ErrBusy, got %v", err)
	}
	down := NewRemote("http://127.0.0.1:1", "x")
	if down.Status(context.Background()).Available {
		t.Error("unreachable daemon must be unavailable")
	}
	if _, err := down.Run(context.Background(), Request{Kind: KindRun}); !errors.Is(err, ErrUnavailable) {
		t.Errorf("want ErrUnavailable, got %v", err)
	}
}

func TestDockerRuntimeFlag(t *testing.T) {
	d := NewDocker(DockerConfig{Bin: "docker", Image: "img", MemoryMB: 512, CPUs: "1", PIDs: 8, Concurrency: 1, Runtime: "runsc"})
	found := false
	args := d.Args("n")
	for i, a := range args {
		if a == "--runtime" && args[i+1] == "runsc" {
			found = true
		}
	}
	if !found || args[len(args)-1] != "img" {
		t.Errorf("runtime flag missing or image not last: %v", args)
	}
}
