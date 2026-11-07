package fetch

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestManifestGzipAndErrors(t *testing.T) {
	body := bytes.Repeat([]byte("<component id=\"x\"/>\n"), 1000)
	s := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/components.xml":
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				http.Error(w, "want gzip", 400)
				return
			}
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			gz.Write(body)
			gz.Close()
		case "/busy":
			w.WriteHeader(503)
		default:
			http.NotFound(w, r)
		}
	})
	c := newClient(t, nil)
	ctx := context.Background()
	got, err := c.FetchManifest(ctx, s.URL+"/components.xml")
	if err != nil || !bytes.Equal(got, body) {
		t.Fatalf("gzip manifest: %v", err)
	}
	if _, err := c.FetchSmall(ctx, s.URL+"/components.xml", 100); err == nil || !strings.Contains(err.Error(), "larger than 100 bytes") {
		t.Fatalf("cap: %v", err)
	}
	if _, err := c.FetchSmall(ctx, s.URL+"/components.xml.minisig", 1<<10); !errors.Is(err, ErrNotFound) || retryable(err) {
		t.Fatalf("404: %v", err)
	}
	if _, err := c.FetchManifest(ctx, s.URL+"/busy"); !retryable(err) || errors.Is(err, ErrNotFound) {
		t.Fatalf("503: %v", err)
	}
	// Connection refused is retryable.
	dead := httptest.NewServer(http.NotFoundHandler())
	dead.Close()
	if _, err := c.FetchSmall(ctx, dead.URL+"/x", 10); !retryable(err) {
		t.Fatalf("refused: %v", err)
	}
	// ctx deadline bounds the manifest request (fpm list budget).
	slow := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	})
	tctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := c.FetchManifest(tctx, slow.URL+"/m.xml"); err == nil || time.Since(start) > time.Second {
		t.Fatalf("deadline: %v after %v", err, time.Since(start))
	}
}
