package fetch

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// validatorExt names the sidecar holding the If-Range validator of a .part.
const validatorExt = ".validator"

// CacheName is the file name FetchArchive uses inside dir: "<crc32>-<Name>",
// so a cached old version never passes for a new one. Names derive from
// validated IDs; this only refuses what is not a plain file name.
func CacheName(it Item) (string, error) {
	if it.Name == "" || it.Name != filepath.Base(it.Name) || !filepath.IsLocal(it.Name) || strings.ContainsAny(it.Name, `/\:`) {
		return "", fmt.Errorf("invalid archive name %q", it.Name)
	}
	return fmt.Sprintf("%08x-%s", it.CRC32, it.Name), nil
}

// FetchArchive makes a verified copy of it in dir and returns its path. A
// cached file that verifies costs no request. Otherwise the archive is
// downloaded into "<path>.part" (resumed when possible; kept on cancel) and
// renamed into place once size and hashes match. Metadata-only items return
// ("", nil). Callbacks may be nil.
//
// Retries: every failure moves on to the next mirror. A mirror whose answer
// is final (404, 403, a different file) is dropped; temporary failures back
// off with jitter. FetchArchive gives up once RetryBudget has passed without
// the .part growing, so the caller can ask the user.
func (c *Client) FetchArchive(ctx context.Context, it Item, dir string, onProgress func(Progress), onRetry func(RetryInfo)) (string, error) {
	if it.Size <= 0 {
		return "", nil
	}
	fail := func(err error) (string, error) { return "", fmt.Errorf("download %s: %w", it.ID, err) }
	name, err := CacheName(it)
	if err != nil {
		return fail(err)
	}
	if len(c.mirrors) == 0 {
		return fail(errors.New("no download mirror configured"))
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fail(fmt.Errorf("create cache folder: %w", err))
	}
	final := filepath.Join(dir, name)
	part := final + ".part"
	prog := &progress{fn: onProgress, id: it.ID, total: it.Size}
	if ok, err := verifyFile(ctx, final, it); err != nil {
		return fail(err)
	} else if ok {
		prog.report(it.Size, true)
		return final, nil
	}
	if os.Remove(final) == nil {
		c.log.Info("cached archive failed verification, downloading again", "id", it.ID)
	}

	var (
		mirrors = slices.Clone(c.mirrors)
		mi      int
		step    int          // backoff step since the last progress
		since   = time.Now() // last progress
		best    int64        // most bytes the .part ever held
	)
	for attempt := 1; ; attempt++ {
		mirror := mirrors[mi]
		resumed := fileSize(part) > 0
		err := c.attempt(ctx, mirror, it, final, prog)
		if err == nil {
			prog.report(it.Size, true)
			return final, nil
		}
		if ctx.Err() != nil {
			return fail(ctx.Err())
		}
		// Progress: the .part grew past its best by 1% of the file (at most
		// 1 MiB), so a trickle does not hold the budget open forever.
		if n := fileSize(part); n >= best+max(1, min(1<<20, it.Size/100)) {
			best, step, since = n, 0, time.Now()
		}
		var wait time.Duration
		switch {
		// A mismatch after resuming may come from a stale prefix; the .part
		// is gone, so the next try starts fresh.
		case retryable(err), resumed && errors.Is(err, ErrHashMismatch):
			step++
			wait = c.backoff(step)
			mi++
		// A final answer (404, 403, a different file): drop this mirror.
		case errors.Is(err, ErrHashMismatch), errors.Is(err, ErrSizeMismatch),
			errors.As(err, new(*statusError)), errors.Is(err, errRedirect):
			mirrors = slices.Delete(mirrors, mi, mi+1)
			if len(mirrors) == 0 {
				if errors.Is(err, ErrNotFound) {
					return fail(fmt.Errorf("%s is not on any mirror (the component list may be out of date): %w", it.Name, err))
				}
				return fail(fmt.Errorf("every mirror failed: %w", err))
			}
		default:
			return fail(err) // local problem (disk full, permissions): retrying won't help
		}
		mi %= len(mirrors)
		left := c.opts.RetryBudget - time.Since(since)
		if left <= 0 {
			return fail(fmt.Errorf("giving up after %d attempts, no progress for %s: %w", attempt, c.opts.RetryBudget, err))
		}
		wait = min(wait, left)
		c.log.Warn("download failed, retrying", "id", it.ID, "attempt", attempt, "wait", wait, "err", err)
		if onRetry != nil {
			onRetry(RetryInfo{ID: it.ID, Attempt: attempt, Wait: wait, Mirror: mirror, Err: err})
		}
		select {
		case <-ctx.Done():
			return fail(ctx.Err())
		case <-time.After(wait):
		}
	}
}

// attempt is one try against one mirror: resume the .part if possible,
// stream the rest into it while hashing, verify, rename.
func (c *Client) attempt(ctx context.Context, base string, it Item, final string, prog *progress) error {
	part := final + ".part"
	have, validator, h, err := resumable(ctx, part, it)
	if err != nil {
		return err
	}
	if have == it.Size { // complete .part, left by a crash or a failed rename
		if h.check() == nil {
			os.Remove(part + validatorExt)
			return renameRetry(part, final)
		}
		removePart(part)
		have, h = 0, newHasher(it)
	}

	// Stall watchdog, started after hashing the prefix; also covers
	// connecting and waiting for headers.
	actx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	stall := time.AfterFunc(c.opts.StallTimeout, func() { cancel(ErrStalled) })
	defer stall.Stop()

	u := base + url.PathEscape(it.Name)
	if it.CacheBust != "" {
		u += "?h=" + url.QueryEscape(it.CacheBust)
	}
	hdr := http.Header{"Accept-Encoding": {"identity"}} // byte offsets must be raw
	if have > 0 {
		hdr.Set("Range", fmt.Sprintf("bytes=%d-", have))
		hdr.Set("If-Range", validator)
	}
	resp, err := c.get(actx, u, hdr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK: // no Range sent, or the file changed since
		if resp.ContentLength >= 0 && resp.ContentLength != it.Size {
			removePart(part)
			return fmt.Errorf("%s: %w: server has %d bytes, the component list says %d", u, ErrSizeMismatch, resp.ContentLength, it.Size)
		}
		have, h = 0, newHasher(it)
	case http.StatusPartialContent:
		var start, end, total int64
		cr := resp.Header.Get("Content-Range")
		if _, err := fmt.Sscanf(cr, "bytes %d-%d/%d", &start, &end, &total); err != nil || start != have || total != it.Size {
			removePart(part)
			return transientError{fmt.Errorf("%s: unexpected Content-Range %q when resuming at byte %d", u, cr, have)}
		}
	case http.StatusRequestedRangeNotSatisfiable:
		removePart(part)
		return transientError{fmt.Errorf("%s: server refused to resume at byte %d", u, have)}
	default:
		io.CopyN(io.Discard, resp.Body, 4<<10) // lets the connection be reused
		return newStatusError(resp, u)
	}

	f, err := os.OpenFile(part, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	if err = f.Truncate(have); err == nil {
		_, err = f.Seek(have, io.SeekStart)
	}
	if err == nil && have == 0 {
		// New bytes need the new validator. If-Range only works with a
		// strong ETag or Last-Modified; with neither, no resume.
		os.Remove(part + validatorExt)
		v := resp.Header.Get("ETag")
		if v == "" || strings.HasPrefix(v, "W/") {
			v = resp.Header.Get("Last-Modified")
		}
		if v != "" {
			err = os.WriteFile(part+validatorExt, []byte(v), 0o644)
		}
	}
	if err != nil {
		f.Close()
		return err
	}

	// Read exactly the missing bytes; anything a server sends beyond
	// the expected size is ignored (the hashes decide).
	prog.start(have)
	done, buf := have, make([]byte, 64<<10)
	var rerr, werr error
	for done < it.Size && rerr == nil && werr == nil {
		var n int
		n, rerr = resp.Body.Read(buf[:min(int64(len(buf)), it.Size-done)])
		stall.Reset(c.opts.StallTimeout)
		_, werr = f.Write(buf[:n])
		h.Write(buf[:n])
		done += int64(n)
		prog.report(done, false)
	}
	switch {
	case werr != nil:
		err = werr
	case done < it.Size && rerr != io.EOF:
		err = fmt.Errorf("%s: %w", u, c.netErr(actx, rerr))
	case done < it.Size:
		err = transientError{fmt.Errorf("%s: connection closed after %d of %d bytes", u, done, it.Size)}
	default:
		if err = h.check(); err != nil {
			err = fmt.Errorf("%s: %w", u, err)
		} else {
			err = f.Sync()
		}
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if errors.Is(err, ErrHashMismatch) {
		removePart(part)
	}
	if err != nil {
		return err
	}
	os.Remove(part + validatorExt)
	return renameRetry(part, final)
}

// resumable returns how many bytes of part can be resumed, a hasher fed with
// them, and the If-Range validator. Unusable leftovers are deleted.
func resumable(ctx context.Context, part string, it Item) (int64, string, *hasher, error) {
	h := newHasher(it)
	fi, err := os.Stat(part)
	v, _ := os.ReadFile(part + validatorExt)
	if err != nil || fi.Size() == 0 || fi.Size() > it.Size || len(v) == 0 && fi.Size() < it.Size {
		removePart(part)
		return 0, "", h, nil
	}
	if err := hashFile(ctx, part, fi.Size(), h); err != nil {
		if ctx.Err() != nil {
			return 0, "", nil, ctx.Err() // keep the .part
		}
		removePart(part)
		return 0, "", newHasher(it), nil
	}
	return fi.Size(), string(v), h, nil
}

func removePart(part string) {
	os.Remove(part)
	os.Remove(part + validatorExt)
}

// backoff is full-jitter exponential backoff, so many clients retrying at
// once spread out instead of hitting the server together.
func (c *Client) backoff(step int) time.Duration {
	return rand.N(max(1, min(c.opts.MaxBackoff, c.opts.BaseBackoff<<min(step-1, 20)))) + 1
}

func fileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}

// hasher computes the CRC32 and, when one is expected, the SHA256.
type hasher struct {
	io.Writer
	it  Item
	crc hash.Hash32
	sha hash.Hash
}

func newHasher(it Item) *hasher {
	h := &hasher{it: it, crc: crc32.NewIEEE()}
	h.Writer = h.crc
	if it.SHA256 != nil {
		h.sha = sha256.New()
		h.Writer = io.MultiWriter(h.crc, h.sha)
	}
	return h
}

func (h *hasher) check() error {
	if got := h.crc.Sum32(); got != h.it.CRC32 {
		return fmt.Errorf("%w: crc32 is %08x, expected %08x (corrupt download or outdated component list)", ErrHashMismatch, got, h.it.CRC32)
	}
	if h.sha != nil && !bytes.Equal(h.sha.Sum(nil), h.it.SHA256) {
		return fmt.Errorf("%w: sha256 is %x, expected %x (corrupt download or outdated component list)", ErrHashMismatch, h.sha.Sum(nil), h.it.SHA256)
	}
	return nil
}

// verifyFile reports whether path holds exactly it. Missing or wrong-size
// files return false without hashing.
func verifyFile(ctx context.Context, path string, it Item) (bool, error) {
	if fi, err := os.Stat(path); err != nil || fi.Size() != it.Size {
		return false, nil
	}
	h := newHasher(it)
	if err := hashFile(ctx, path, it.Size, h); err != nil {
		return false, ctx.Err()
	}
	return h.check() == nil, nil
}

// hashFile feeds the first n bytes of path to h, checking ctx between chunks.
func hashFile(ctx context.Context, path string, n int64, h *hasher) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := make([]byte, 256<<10)
	for n > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		k, err := io.ReadFull(f, buf[:min(n, int64(len(buf)))])
		h.Write(buf[:k])
		if err != nil {
			return err
		}
		n -= int64(k)
	}
	return nil
}

// progress throttles callbacks to ~5 Hz.
type progress struct {
	fn          func(Progress)
	id          string
	total, from int64     // from: bytes present when the attempt started
	t0, last    time.Time // attempt start, last report
}

func (p *progress) start(done int64) {
	p.from, p.t0 = done, time.Now()
	p.report(done, true)
}

func (p *progress) report(done int64, force bool) {
	if p.fn == nil || !force && time.Since(p.last) < 200*time.Millisecond {
		return
	}
	p.last = time.Now()
	var rate float64
	if el := p.last.Sub(p.t0).Seconds(); !p.t0.IsZero() && el > 0 {
		rate = float64(done-p.from) / el
	}
	p.fn(Progress{ID: p.id, Done: done, Total: p.total, BytesPerSec: rate})
}
