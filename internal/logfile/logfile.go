// Package logfile is the log file of fpm and the manager:
// <root>/Logs/<name>.log, appended to, moved to <name>.log.1 once it passes
// 5 MB.
package logfile

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
)

const (
	rotateSize = 5 << 20   // a log at least this big is moved to path.1 on Open
	earlyLimit = 256 << 10 // most bytes kept before Open
)

// Sink is an io.Writer for a slog handler. It keeps lines in memory until
// Open is called (the log lives in the Flashpoint folder, which is found
// only after the program has started logging), then writes them and
// everything after to the file. Logging is best effort: Write never fails,
// so a read-only Flashpoint folder cannot break `fpm list`, and a writer
// combined with it in an io.MultiWriter (stderr) keeps working.
type Sink struct {
	mu    sync.Mutex
	f     *os.File
	path  string
	early bytes.Buffer
}

func (s *Sink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.f != nil {
		s.f.Write(p)
	} else if s.early.Len()+len(p) <= earlyLimit {
		s.early.Write(p)
	}
	return len(p), nil
}

// Open switches to the log file at path (again after the manager's
// Settings change the Flashpoint folder), first moving a log over 5 MB to
// path.1. That rename fails while another fpm or manager has the log open
// on Windows; the log then keeps growing until the next start. On error
// the sink keeps its previous file (or keeps buffering).
func (s *Sink) Open(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if path == s.path {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if fi, err := os.Stat(path); err == nil && fi.Size() >= rotateSize {
		os.Rename(path, path+".1")
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	if s.f != nil {
		s.f.Close()
	}
	f.Write(s.early.Bytes())
	s.early.Reset()
	s.f, s.path = f, path
	return nil
}

// Close closes the file; later writes are buffered again.
func (s *Sink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.f == nil {
		return nil
	}
	err := s.f.Close()
	s.f, s.path = nil, ""
	return err
}
