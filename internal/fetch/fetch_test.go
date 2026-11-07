package fetch

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---- helpers ----

func testData(n int, seed uint64) []byte {
	r := rand.New(rand.NewPCG(seed, seed))
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(r.Uint32())
	}
	return b
}

func itemFor(name string, data []byte) Item {
	sum := sha256.Sum256(data)
	return Item{ID: strings.TrimSuffix(name, ".zip"), Name: name, Size: int64(len(data)),
		CRC32: crc32.ChecksumIEEE(data), SHA256: sum[:]}
}

func etagOf(data []byte) string {
	s := sha1.Sum(data)
	return `"` + hex.EncodeToString(s[:]) + `"`
}

// serve answers like Nexus: strong sha1 ETag, Range + If-Range support.
func serve(w http.ResponseWriter, r *http.Request, data []byte) {
	w.Header().Set("ETag", etagOf(data))
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data))
}

// dropAfter sends full headers and n body bytes, then closes the connection.
func dropAfter(w http.ResponseWriter, data []byte, n int) {
	conn, buf, err := w.(http.Hijacker).Hijack()
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(buf, "HTTP/1.1 200 OK\r\nContent-Length: %d\r\nETag: %s\r\n\r\n", len(data), etagOf(data))
	buf.Write(data[:n])
	buf.Flush()
	conn.Close()
}

type server struct {
	*httptest.Server
	hits   atomic.Int64
	mu     sync.Mutex
	ranges []string // Range header of every request
}

func newServer(t *testing.T, h func(n int64, w http.ResponseWriter, r *http.Request)) *server {
	s := &server{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := s.hits.Add(1)
		s.mu.Lock()
		s.ranges = append(s.ranges, r.Header.Get("Range"))
		s.mu.Unlock()
		h(n, w, r)
	}))
	t.Cleanup(s.Close)
	return s
}

func (s *server) rangeOf(i int) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ranges[i]
}

func newClient(t *testing.T, mod func(*Options), mirrors ...string) *Client {
	o := Options{
		Mirrors:      mirrors,
		BaseBackoff:  time.Millisecond,
		MaxBackoff:   20 * time.Millisecond,
		StallTimeout: 2 * time.Second,
		Logger:       slog.New(slog.NewTextHandler(t.Output(), &slog.HandlerOptions{Level: slog.LevelWarn})),
	}
	if mod != nil {
		mod(&o)
	}
	return New(o)
}

func mustFile(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s: content differs (%d vs %d bytes)", path, len(got), len(want))
	}
}

func noFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("%s should not exist (err=%v)", path, err)
	}
}

// ---- archives ----

func TestHappyPath(t *testing.T) {
	data := testData(300_000, 1)
	it := itemFor("core-x.zip", data)
	it.CacheBust = "a b"
	s := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repo/core-x.zip" || r.URL.Query().Get("h") != "a b" || r.Header.Get("User-Agent") != "test-ua" {
			http.Error(w, "bad request "+r.URL.String(), 400)
			return
		}
		serve(w, r, data)
	})
	c := newClient(t, func(o *Options) { o.UserAgent = "test-ua" }, s.URL+"/repo") // no trailing slash on purpose
	dir := t.TempDir()
	var last Progress
	path, err := c.FetchArchive(context.Background(), it, dir, func(p Progress) { last = p }, nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, fmt.Sprintf("%08x-core-x.zip", it.CRC32)); path != want {
		t.Fatalf("path %q, want %q", path, want)
	}
	mustFile(t, path, data)
	noFile(t, path+".part")
	noFile(t, path+".part"+validatorExt)
	if last.Done != it.Size || last.Total != it.Size || last.ID != "core-x" {
		t.Fatalf("last progress %+v", last)
	}
}

func TestCacheHit(t *testing.T) {
	data := testData(50_000, 2)
	it := itemFor("a.zip", data)
	s := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) { serve(w, r, data) })
	c := newClient(t, nil, s.URL+"/")
	dir := t.TempDir()
	ctx := context.Background()
	path, err := c.FetchArchive(ctx, it, dir, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.FetchArchive(ctx, it, dir, nil, nil); err != nil {
		t.Fatal(err)
	}
	if n := s.hits.Load(); n != 1 {
		t.Fatalf("cache hit made a request: %d hits", n)
	}
	// A corrupted cached file is detected and replaced.
	os.WriteFile(path, bytes.Repeat([]byte{0}, len(data)), 0o644)
	if _, err := c.FetchArchive(ctx, it, dir, nil, nil); err != nil {
		t.Fatal(err)
	}
	mustFile(t, path, data)
	if n := s.hits.Load(); n != 2 {
		t.Fatalf("hits %d, want 2", n)
	}
	// A complete .part left by a crash before rename is used without network.
	os.Rename(path, path+".part")
	if _, err := c.FetchArchive(ctx, it, dir, nil, nil); err != nil {
		t.Fatal(err)
	}
	mustFile(t, path, data)
	if n := s.hits.Load(); n != 2 {
		t.Fatalf("hits %d, want 2", n)
	}
}

func TestMetadataOnly(t *testing.T) {
	c := newClient(t, nil, "http://127.0.0.1:1/")
	path, err := c.FetchArchive(context.Background(), Item{ID: "x", Name: "x.zip"}, t.TempDir(), nil, nil)
	if path != "" || err != nil {
		t.Fatalf("got %q, %v", path, err)
	}
}

func TestResume(t *testing.T) {
	data := testData(200_000, 3)
	it := itemFor("r.zip", data)
	const cut = 70_000
	tests := []struct {
		name     string
		second   func(w http.ResponseWriter, r *http.Request)
		wantHits int64 // 416 or a bad Content-Range: one more, from scratch
	}{
		{"206 resume", func(w http.ResponseWriter, r *http.Request) { serve(w, r, data) }, 2},
		{"server ignores Range", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("ETag", etagOf(data))
			w.Write(data)
		}, 2},
		{"416 then full", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Range") != "" {
				w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				return
			}
			serve(w, r, data)
		}, 3},
		{"wrong Content-Range start", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Range") != "" {
				w.Header().Set("Content-Range", fmt.Sprintf("bytes 0-%d/%d", len(data)-1, len(data)))
				w.WriteHeader(http.StatusPartialContent)
				w.Write(data)
				return
			}
			serve(w, r, data)
		}, 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newServer(t, func(n int64, w http.ResponseWriter, r *http.Request) {
				if n == 1 {
					dropAfter(w, data, cut)
					return
				}
				tc.second(w, r)
			})
			var retries []RetryInfo
			c := newClient(t, nil, s.URL+"/")
			path, err := c.FetchArchive(context.Background(), it, t.TempDir(), nil, func(ri RetryInfo) { retries = append(retries, ri) })
			if err != nil {
				t.Fatal(err)
			}
			mustFile(t, path, data)
			if got := s.hits.Load(); got != tc.wantHits {
				t.Fatalf("hits %d, want %d", got, tc.wantHits)
			}
			if got := s.rangeOf(1); got != "bytes=70000-" {
				t.Fatalf("2nd request Range %q", got)
			}
			if len(retries) != int(tc.wantHits-1) || !retryable(retries[0].Err) {
				t.Fatalf("retries %+v", retries)
			}
		})
	}
}

func TestIfRangeChangedETag(t *testing.T) {
	data := testData(100_000, 4)
	it := itemFor("e.zip", data)
	dir := t.TempDir()
	name, _ := CacheName(it)
	part := filepath.Join(dir, name) + ".part"
	// Leftover from an older upload: different bytes, different ETag.
	os.WriteFile(part, testData(40_000, 99), 0o644)
	os.WriteFile(part+validatorExt, []byte(`"old"`), 0o644)
	s := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-Range") != `"old"` {
			http.Error(w, "missing If-Range", 400)
			return
		}
		serve(w, r, data) // ETag differs, so ServeContent sends 200 with the whole file
	})
	c := newClient(t, nil, s.URL+"/")
	path, err := c.FetchArchive(context.Background(), it, dir, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	mustFile(t, path, data)
	if s.hits.Load() != 1 || s.rangeOf(0) != "bytes=40000-" {
		t.Fatalf("hits %d range %q", s.hits.Load(), s.rangeOf(0))
	}
}

func TestHashMismatchFailover(t *testing.T) {
	data := testData(30_000, 5)
	it := itemFor("h.zip", data)
	bad := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) { serve(w, r, testData(len(data), 6)) })
	good := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) { serve(w, r, data) })

	var retries []RetryInfo
	c := newClient(t, nil, bad.URL+"/", good.URL+"/")
	path, err := c.FetchArchive(context.Background(), it, t.TempDir(), nil, func(ri RetryInfo) { retries = append(retries, ri) })
	if err != nil {
		t.Fatal(err)
	}
	mustFile(t, path, data)
	if bad.hits.Load() != 1 || good.hits.Load() != 1 {
		t.Fatalf("hits bad=%d good=%d, want 1 and 1", bad.hits.Load(), good.hits.Load())
	}
	if len(retries) != 1 || !errors.Is(retries[0].Err, ErrHashMismatch) || retries[0].Wait != 0 {
		t.Fatalf("retries %+v", retries)
	}

	// Only bad mirrors: each is tried once, typed error.
	c = newClient(t, nil, bad.URL+"/", bad.URL+"/again/")
	dir := t.TempDir()
	_, err = c.FetchArchive(context.Background(), it, dir, nil, nil)
	if !errors.Is(err, ErrHashMismatch) || !strings.Contains(err.Error(), "crc32") || retryable(err) {
		t.Fatalf("err %v", err)
	}
	if bad.hits.Load() != 3 {
		t.Fatalf("bad hits %d, want 3", bad.hits.Load())
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatalf("leftover files: %v", entries)
	}
}

func TestNotFound(t *testing.T) {
	data := testData(1000, 7)
	it := itemFor("n.zip", data)
	missing := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	good := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) { serve(w, r, data) })

	c := newClient(t, nil, missing.URL+"/", good.URL+"/")
	if _, err := c.FetchArchive(context.Background(), it, t.TempDir(), nil, nil); err != nil {
		t.Fatal(err)
	}
	c = newClient(t, nil, missing.URL+"/", missing.URL+"/b/")
	_, err := c.FetchArchive(context.Background(), it, t.TempDir(), nil, nil)
	if !errors.Is(err, ErrNotFound) || retryable(err) || !strings.Contains(err.Error(), "n.zip") {
		t.Fatalf("err %v", err)
	}
	if missing.hits.Load() != 3 {
		t.Fatalf("404 mirror hits %d, want 3 (no retry on same mirror)", missing.hits.Load())
	}
}

func TestBackoff(t *testing.T) {
	c := New(Options{BaseBackoff: time.Second, MaxBackoff: 8 * time.Second})
	for n := 1; n < 70; n++ {
		limit := min(time.Second<<min(n-1, 10), 8*time.Second)
		if d := c.backoff(n); d <= 0 || d > limit {
			t.Fatalf("backoff(%d) = %v, limit %v", n, d, limit)
		}
	}
}

func TestStall(t *testing.T) {
	data := testData(50_000, 9)
	it := itemFor("s.zip", data)
	stallFirst := func(n int64, w http.ResponseWriter, r *http.Request) {
		if n > 1 && r.Header.Get("Range") != "" {
			serve(w, r, data)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprint(len(data)))
		w.Header().Set("ETag", etagOf(data))
		w.Write(data[:10])
		w.(http.Flusher).Flush()
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}
	s := newServer(t, stallFirst)
	var retries []RetryInfo
	c := newClient(t, func(o *Options) { o.StallTimeout = 100 * time.Millisecond }, s.URL+"/")
	start := time.Now()
	path, err := c.FetchArchive(context.Background(), it, t.TempDir(), nil, func(ri RetryInfo) { retries = append(retries, ri) })
	if err != nil {
		t.Fatal(err)
	}
	mustFile(t, path, data)
	if el := time.Since(start); el > 2*time.Second {
		t.Fatalf("stall detection too slow: %v", el)
	}
	if len(retries) != 1 || !errors.Is(retries[0].Err, ErrStalled) || s.rangeOf(1) != "bytes=10-" {
		t.Fatalf("retries %+v range %q", retries, s.rangeOf(1))
	}

	// Stalls every time: gives up with ErrStalled once the budget is spent.
	s2 := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) { stallFirst(1, w, r) })
	c = newClient(t, func(o *Options) { o.StallTimeout = 100 * time.Millisecond; o.RetryBudget = 300 * time.Millisecond }, s2.URL+"/")
	start = time.Now()
	if _, err := c.FetchArchive(context.Background(), it, t.TempDir(), nil, nil); !errors.Is(err, ErrStalled) || !retryable(err) || time.Since(start) > 2*time.Second {
		t.Fatalf("err %v after %v", err, time.Since(start))
	}
}

func TestCancelKeepsPart(t *testing.T) {
	data := testData(100_000, 10)
	it := itemFor("c.zip", data)
	sent := make(chan struct{})
	s := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprint(len(data)))
		w.Header().Set("ETag", etagOf(data))
		w.Write(data[:4096])
		w.(http.Flusher).Flush()
		close(sent)
		<-r.Context().Done()
	})
	c := newClient(t, func(o *Options) { o.StallTimeout = 10 * time.Second }, s.URL+"/")
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() {
		_, err := c.FetchArchive(ctx, it, dir, nil, nil)
		errc <- err
	}()
	<-sent
	time.Sleep(50 * time.Millisecond)
	start := time.Now()
	cancel()
	err := <-errc
	if el := time.Since(start); el > 100*time.Millisecond {
		t.Fatalf("cancel took %v", el)
	}
	if !errors.Is(err, context.Canceled) || retryable(err) {
		t.Fatalf("err %v", err)
	}
	name, _ := CacheName(it)
	fi, err := os.Stat(filepath.Join(dir, name+".part"))
	if err != nil || fi.Size() != 4096 {
		t.Fatalf("part: %v size %v", err, fi)
	}
	if _, err := os.Stat(filepath.Join(dir, name+".part"+validatorExt)); err != nil {
		t.Fatal(err)
	}
}

func TestOversize(t *testing.T) {
	data := testData(10_000, 11)
	big := append(bytes.Clone(data), make([]byte, 100)...)
	s := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) { serve(w, r, big) })
	dir := t.TempDir()
	_, err := newClient(t, nil, s.URL+"/").FetchArchive(context.Background(), itemFor("o.zip", data), dir, nil, nil)
	if !errors.Is(err, ErrSizeMismatch) || retryable(err) || s.hits.Load() != 1 {
		t.Fatalf("hits %d (want 1: a different file is a final answer), err %v", s.hits.Load(), err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("leftover files: %v", entries)
	}
}

// A weak ETag cannot be used in If-Range: without Last-Modified the
// interrupted download restarts from 0 instead of resuming.
func TestWeakETagNoResume(t *testing.T) {
	data := testData(100_000, 13)
	s := newServer(t, func(n int64, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `W/"weak"`)
		if n > 1 {
			w.Write(data)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprint(len(data)))
		w.Write(data[:30_000])
		w.(http.Flusher).Flush()
		panic(http.ErrAbortHandler)
	})
	path, err := newClient(t, nil, s.URL+"/").FetchArchive(context.Background(), itemFor("w.zip", data), t.TempDir(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	mustFile(t, path, data)
	if s.hits.Load() != 2 || s.rangeOf(1) != "" {
		t.Fatalf("hits %d, 2nd Range %q", s.hits.Load(), s.rangeOf(1))
	}
}

func TestCacheName(t *testing.T) {
	for name, ok := range map[string]bool{
		"core-launcher.zip": true, "a b.zip": true,
		"": false, "..": false, "../x.zip": false, "a/b.zip": false, `a\b.zip`: false,
		"c:x.zip": false, "x.zip:ads": false, "/abs.zip": false,
	} {
		got, err := CacheName(Item{Name: name, CRC32: 0xab})
		if (err == nil) != ok {
			t.Errorf("CacheName(%q) err=%v, want ok=%v", name, err, ok)
		}
		if ok && got != "000000ab-"+name {
			t.Errorf("CacheName(%q) = %q", name, got)
		}
	}
}

// ---- redirects ----

func TestRedirect(t *testing.T) {
	data := testData(2000, 12)
	plain := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) { serve(w, r, data) })
	var tlsHits atomic.Int64
	tlsSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tlsHits.Add(1)
		switch r.URL.Path {
		case "/down/r.zip":
			http.Redirect(w, r, plain.URL+"/r.zip", http.StatusFound)
		case "/hop/r.zip":
			http.Redirect(w, r, "/final/r.zip", http.StatusMovedPermanently)
		default:
			serve(w, r, data)
		}
	}))
	t.Cleanup(tlsSrv.Close)
	it := itemFor("r.zip", data)

	c := newClient(t, func(o *Options) { o.HTTPClient = tlsSrv.Client() }, tlsSrv.URL+"/down/")
	_, err := c.FetchArchive(context.Background(), it, t.TempDir(), nil, nil)
	if !errors.Is(err, errRedirect) || plain.hits.Load() != 0 || tlsHits.Load() != 1 {
		t.Fatalf("err %v plain hits %d tls hits %d", err, plain.hits.Load(), tlsHits.Load())
	}
	if _, err := c.FetchSmall(context.Background(), tlsSrv.URL+"/down/r.zip", 1<<20); !errors.Is(err, errRedirect) {
		t.Fatalf("FetchSmall err %v", err)
	}

	c = newClient(t, func(o *Options) { o.HTTPClient = tlsSrv.Client() }, tlsSrv.URL+"/hop/")
	path, err := c.FetchArchive(context.Background(), it, t.TempDir(), nil, nil)
	if err != nil {
		t.Fatalf("https->https redirect: %v", err)
	}
	mustFile(t, path, data)
}

// ---- manifest and small files ----

// ---- retry policy (progress, time budget) ----

// dropEvery serves data like Nexus but closes the connection after each
// chunk bytes of body, so every attempt makes some progress.
func dropEvery(data []byte, chunk int) func(int64, http.ResponseWriter, *http.Request) {
	return func(_ int64, w http.ResponseWriter, r *http.Request) {
		var start int
		if rg := r.Header.Get("Range"); rg != "" {
			fmt.Sscanf(rg, "bytes=%d-", &start)
		}
		conn, buf, err := w.(http.Hijacker).Hijack()
		if err != nil {
			panic(err)
		}
		defer conn.Close()
		end := min(start+chunk, len(data))
		if start > 0 {
			fmt.Fprintf(buf, "HTTP/1.1 206 Partial Content\r\nContent-Range: bytes %d-%d/%d\r\nContent-Length: %d\r\nETag: %s\r\n\r\n",
				start, len(data)-1, len(data), len(data)-start, etagOf(data))
		} else {
			fmt.Fprintf(buf, "HTTP/1.1 200 OK\r\nContent-Length: %d\r\nETag: %s\r\n\r\n", len(data), etagOf(data))
		}
		buf.Write(data[start:end])
		buf.Flush()
	}
}

func TestProgressDoesNotCount(t *testing.T) {
	// 10 drops at 10% each, each attempt slower than the budget: progress
	// restarts the budget.
	data := testData(200_000, 20)
	it := itemFor("big.zip", data)
	drop := dropEvery(data, 20_000)
	s := newServer(t, func(n int64, w http.ResponseWriter, r *http.Request) {
		time.Sleep(60 * time.Millisecond)
		drop(n, w, r)
	})
	c := newClient(t, func(o *Options) { o.RetryBudget = 50 * time.Millisecond }, s.URL+"/")
	var retries int
	path, err := c.FetchArchive(context.Background(), it, t.TempDir(), nil, func(RetryInfo) { retries++ })
	if err != nil {
		t.Fatal(err)
	}
	mustFile(t, path, data)
	if s.hits.Load() != 10 || retries != 9 {
		t.Fatalf("hits %d retries %d, want 10 and 9", s.hits.Load(), retries)
	}

	// A server that ignores Range restarts from 0 each time: no progress
	// after the first attempt, so the budget runs out.
	noRange := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) {
		r.Header.Del("Range")
		dropAfter(w, data, 20_000)
	})
	c = newClient(t, func(o *Options) { o.RetryBudget = 100 * time.Millisecond }, noRange.URL+"/")
	start := time.Now()
	_, err = c.FetchArchive(context.Background(), it, t.TempDir(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "no progress for 100ms") || noRange.hits.Load() < 3 || time.Since(start) > time.Second {
		t.Fatalf("hits %d err %v after %v", noRange.hits.Load(), err, time.Since(start))
	}
}

func TestRetryBudget(t *testing.T) {
	data := testData(1000, 22)
	it := itemFor("rb.zip", data)

	// A restart: 503 for a while, then fine. Only time decides.
	var upAt atomic.Int64
	upAt.Store(time.Now().Add(150 * time.Millisecond).UnixNano())
	restarting := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) {
		if time.Now().UnixNano() < upAt.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		serve(w, r, data)
	})
	var retries []RetryInfo
	c := newClient(t, func(o *Options) {
		o.RetryBudget = 2 * time.Second
		o.BaseBackoff, o.MaxBackoff = time.Millisecond, 10*time.Millisecond
	}, restarting.URL+"/")
	if _, err := c.FetchArchive(context.Background(), it, t.TempDir(), nil, func(ri RetryInfo) { retries = append(retries, ri) }); err != nil {
		t.Fatalf("did not survive the restart: %v", err)
	}
	var se *statusError
	if len(retries) < 2 || retries[1].Attempt != 2 || !errors.As(retries[0].Err, &se) || se.code != 503 || retries[0].Wait <= 0 {
		t.Fatalf("retries %+v", retries)
	}

	// Down for good: gives up once the budget is spent, naming the cause.
	down := newServer(t, func(_ int64, w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadGateway) })
	c = newClient(t, func(o *Options) { o.RetryBudget = 100 * time.Millisecond }, down.URL+"/")
	start := time.Now()
	_, err := c.FetchArchive(context.Background(), it, t.TempDir(), nil, nil)
	if el := time.Since(start); !errors.As(err, &se) || se.code != 502 || !strings.Contains(err.Error(), "download rb: giving up after") ||
		el < 100*time.Millisecond || el > time.Second {
		t.Fatalf("after %v: %v", el, err)
	}

	// Every failure moves on to the next mirror.
	c = newClient(t, nil, down.URL+"/", restarting.URL+"/")
	before := down.hits.Load()
	if _, err := c.FetchArchive(context.Background(), it, t.TempDir(), nil, nil); err != nil || down.hits.Load() != before+1 {
		t.Fatalf("failover: %v, down hits %d", err, down.hits.Load()-before)
	}

	// A backoff longer than the remaining budget is cut to the budget.
	var waits []time.Duration
	c = newClient(t, func(o *Options) {
		o.RetryBudget = 100 * time.Millisecond
		o.BaseBackoff, o.MaxBackoff = time.Hour, time.Hour
	}, down.URL+"/")
	start = time.Now()
	_, err = c.FetchArchive(context.Background(), it, t.TempDir(), nil, func(ri RetryInfo) { waits = append(waits, ri.Wait) })
	if el := time.Since(start); err == nil || el > time.Second || len(waits) == 0 || waits[0] > 100*time.Millisecond {
		t.Fatalf("after %v: waits %v err %v", el, waits, err)
	}

	// ctx still wins over a long wait.
	c = newClient(t, func(o *Options) { o.RetryBudget = time.Hour; o.BaseBackoff, o.MaxBackoff = time.Hour, time.Hour }, down.URL+"/")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start = time.Now()
	if _, err := c.FetchArchive(ctx, it, t.TempDir(), nil, nil); !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > time.Second {
		t.Fatalf("ctx: %v after %v", err, time.Since(start))
	}
}
