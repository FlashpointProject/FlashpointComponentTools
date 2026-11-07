// Package fetch downloads the manifest and component archives (DESIGN.md
// §1.3, §4.4): one GET for the manifest, one GET per needed zip, resume
// instead of restart, backoff with jitter, small parallelism, stall
// detection. An archive is handed out only after size, CRC32 and SHA256 (if
// known) verified; it is streamed into *.part and renamed into place.
package fetch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// Item is one archive to fetch.
type Item struct {
	ID     string // for logging/progress
	Name   string // file name relative to a mirror base URL, e.g. "core-launcher.zip"
	Size   int64  // expected length. 0 means metadata-only: nothing is fetched.
	CRC32  uint32 // expected CRC32 (IEEE) of the whole file
	SHA256 []byte // verified when non-nil
	// CacheBust is appended as "?h=<CacheBust>" so URL-keyed CDN caches never
	// serve a stale archive for a new manifest.
	CacheBust string
}

// Progress is reported about 5 times a second during a download.
type Progress struct {
	ID          string
	Done        int64 // bytes present, including resumed ones
	Total       int64
	BytesPerSec float64 // average of the current attempt; 0 until known
}

// RetryInfo is reported before sleeping for a retry.
type RetryInfo struct {
	ID      string
	Attempt int           // 1-based number of the attempt that failed
	Wait    time.Duration // backoff before the next attempt
	Mirror  string        // mirror that failed
	Err     error
}

const (
	DefaultUserAgent   = "FlashpointManager/2 (+https://flashpointarchive.org)" // version appended by caller
	DefaultParallel    = 2                                                      // concurrent archive downloads (the engine uses this)
	DefaultRetryBudget = time.Minute                                            // retrying without progress, then the caller asks the user
	MaxManifestBytes   = 16 << 20
)

// Options configure a Client. Zero values mean the defaults.
type Options struct {
	Mirrors   []string // base URLs; first is primary
	UserAgent string
	// RetryBudget is how long one archive keeps retrying since its last
	// progress before FetchArchive returns the error (the GUI then offers
	// Retry/Skip/Abort, which covers a longer Nexus restart).
	RetryBudget  time.Duration
	StallTimeout time.Duration // no data for this long aborts an attempt; default 30 s
	BaseBackoff  time.Duration // first backoff step, doubling; default 1 s
	MaxBackoff   time.Duration // default 1 min
	HTTPClient   *http.Client  // tests; its CheckRedirect is replaced unless set
	Logger       *slog.Logger
}

// Client downloads archives and small files. It is safe for concurrent use,
// but the same Item must not be fetched into the same dir concurrently.
type Client struct {
	opts    Options // defaults filled in
	mirrors []string
	http    *http.Client
	log     *slog.Logger
}

// New builds a Client.
func New(o Options) *Client {
	o.UserAgent = or(o.UserAgent, DefaultUserAgent)
	o.RetryBudget = or(o.RetryBudget, DefaultRetryBudget)
	o.StallTimeout = or(o.StallTimeout, 30*time.Second)
	o.BaseBackoff = or(o.BaseBackoff, time.Second)
	o.MaxBackoff = or(o.MaxBackoff, time.Minute)
	c := &Client{opts: o, log: o.Logger, http: &http.Client{}}
	if c.log == nil {
		c.log = slog.New(slog.DiscardHandler)
	}
	for _, m := range o.Mirrors {
		if m = strings.TrimSpace(m); m != "" {
			c.mirrors = append(c.mirrors, strings.TrimSuffix(m, "/")+"/")
		}
	}
	if o.HTTPClient != nil {
		*c.http = *o.HTTPClient
	}
	if c.http.CheckRedirect == nil {
		c.http.CheckRedirect = checkRedirect
	}
	return c
}

// checkRedirect allows a few redirects (CDN or mirror moves), never from
// https to plain http.
func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		return fmt.Errorf("%w: too many redirects", errRedirect)
	}
	if via[0].URL.Scheme == "https" && req.URL.Scheme != "https" {
		return fmt.Errorf("%w: %s redirects to insecure %s", errRedirect, via[len(via)-1].URL.Redacted(), req.URL.Redacted())
	}
	return nil
}

func (c *Client) get(ctx context.Context, u string, hdr http.Header) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("URL %q: %w", u, err)
	}
	for k, v := range hdr {
		req.Header[k] = v
	}
	req.Header.Set("User-Agent", c.opts.UserAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, c.netErr(ctx, err)
	}
	return resp, nil
}

var (
	ErrNotFound     = errors.New("not found on server") // 404/410 (from every mirror, for archives)
	ErrHashMismatch = errors.New("checksum mismatch")
	ErrSizeMismatch = errors.New("size mismatch") // server's size differs from the manifest's
	ErrStalled      = errors.New("download stalled")
	errRedirect     = errors.New("redirect refused")
)

// statusError is an unexpected HTTP status. 404/410 match ErrNotFound.
type statusError struct {
	url  string
	code int
}

func newStatusError(resp *http.Response, u string) *statusError {
	return &statusError{url: u, code: resp.StatusCode}
}

func (e *statusError) Error() string {
	return fmt.Sprintf("%s: server returned HTTP %d %s", e.url, e.code, http.StatusText(e.code))
}

func (e *statusError) Is(target error) bool {
	return target == ErrNotFound && (e.code == http.StatusNotFound || e.code == http.StatusGone)
}

// transientError marks network failures (reset, refused, EOF, timeout).
type transientError struct{ error }

func (e transientError) Unwrap() error { return e.error }

// retryable reports whether trying again later may succeed: network errors,
// stalls and temporary HTTP statuses.
func retryable(err error) bool {
	var se *statusError
	if errors.As(err, &se) {
		switch se.code {
		case 408, 425, 429, 500, 502, 503, 504:
			return true
		}
		return false
	}
	return errors.Is(err, ErrStalled) || errors.As(err, new(transientError)) && !errors.Is(err, context.Canceled)
}

// netErr classifies an error from a request or a body read.
func (c *Client) netErr(ctx context.Context, err error) error {
	if errors.Is(context.Cause(ctx), ErrStalled) {
		return fmt.Errorf("%w: no data for %s", ErrStalled, c.opts.StallTimeout)
	}
	if errors.Is(err, errRedirect) {
		return err
	}
	return transientError{err}
}

// renameRetry is os.Rename, retried for a few seconds on Windows: antivirus
// scanners open a freshly written file for a moment, and Windows refuses to
// rename it meanwhile.
func renameRetry(from, to string) error {
	deadline := time.Now().Add(3 * time.Second)
	for delay := 10 * time.Millisecond; ; delay = min(delay*2, 500*time.Millisecond) {
		err := os.Rename(from, to)
		if err == nil || !isTransientFSError(err) || time.Now().Add(delay).After(deadline) {
			return err
		}
		time.Sleep(delay)
	}
}

func or[T comparable](v, def T) T {
	var zero T
	if v == zero {
		return def
	}
	return v
}
