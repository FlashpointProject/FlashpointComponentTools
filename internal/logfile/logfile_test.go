package logfile

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSink(t *testing.T) {
	var s Sink
	s.Write([]byte("early\n"))
	p := filepath.Join(t.TempDir(), "Logs", "manager.log")
	if err := s.Open(p); err != nil {
		t.Fatal(err)
	}
	s.Write([]byte("late\n"))
	other := filepath.Join(t.TempDir(), "Logs", "manager.log")
	if err := s.Open(other); err != nil {
		t.Fatal(err)
	}
	s.Write([]byte("moved\n"))
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(p); string(b) != "early\nlate\n" {
		t.Errorf("first log = %q", b)
	}
	if b, _ := os.ReadFile(other); string(b) != "moved\n" {
		t.Errorf("second log = %q", b)
	}

	// Appends to an existing log.
	var again Sink
	again.Write([]byte("next run\n"))
	if err := again.Open(other); err != nil {
		t.Fatal(err)
	}
	again.Close()
	if b, _ := os.ReadFile(other); string(b) != "moved\nnext run\n" {
		t.Errorf("appended log = %q", b)
	}
}

func TestRotate(t *testing.T) {
	p := filepath.Join(t.TempDir(), "fpm.log")
	big := bytes.Repeat([]byte("x"), rotateSize)
	if err := os.WriteFile(p, big, 0o644); err != nil {
		t.Fatal(err)
	}
	var s Sink
	if err := s.Open(p); err != nil {
		t.Fatal(err)
	}
	s.Write([]byte("new\n"))
	s.Close()
	if b, _ := os.ReadFile(p); string(b) != "new\n" {
		t.Errorf("log after rotation = %q", b)
	}
	if fi, err := os.Stat(p + ".1"); err != nil || fi.Size() != rotateSize {
		t.Errorf("rotated log: %v", err)
	}
}

func TestOpenFailsBuffers(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	os.WriteFile(file, nil, 0o644)
	var s Sink
	s.Write([]byte("kept\n"))
	if err := s.Open(filepath.Join(file, "Logs", "fpm.log")); err == nil {
		t.Fatal("Open under a file succeeded")
	}
	if n, err := s.Write([]byte("more\n")); n != 5 || err != nil {
		t.Errorf("Write = %d, %v", n, err)
	}
	p := filepath.Join(t.TempDir(), "fpm.log")
	s.Open(p)
	s.Close()
	if b, _ := os.ReadFile(p); string(b) != "kept\nmore\n" {
		t.Errorf("log = %q", b)
	}
}
