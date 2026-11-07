package state

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/manifest"
)

// ErrCorrupt is wrapped by every error about an unparsable record.
var ErrCorrupt = errors.New("corrupt component record")

// RecordError reports a record that could not be read or parsed. The
// component should be treated as installed but broken (offer a reinstall).
type RecordError struct {
	ID   string // file name in Components/
	Path string
	Err  error
}

func (e *RecordError) Error() string {
	return fmt.Sprintf("component record %q (%s): %v", e.ID, e.Path, e.Err)
}

func (e *RecordError) Unwrap() error { return e.Err }

// ParseRecord parses the content of Components/<id>. It accepts what any
// writer produced: CRLF or LF, a UTF-8 BOM, trailing whitespace, blank lines
// and either slash direction. Only a bad ID or header is an error (wrapping
// ErrCorrupt). A path line that fails manifest.CleanRelPath is skipped with
// a warning in r.Warnings, because uninstall deletes exactly these paths;
// the component itself still counts as installed, as for the legacy tools.
func ParseRecord(id string, data []byte) (*Record, error) {
	if !manifest.ValidID(id) {
		return nil, fmt.Errorf("%w: %q is not a valid component ID", ErrCorrupt, id)
	}
	lines := strings.Split(string(bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))), "\n")
	header := strings.Fields(lines[0]) // C# takes the first line as the header
	if len(header) < 2 {
		return nil, fmt.Errorf("%w: header %q is not \"<hash> <size> [deps...]\"", ErrCorrupt, strings.TrimSpace(lines[0]))
	}
	size, err := strconv.ParseInt(header[1], 10, 64)
	if err != nil || size < 0 || header[1][0] == '+' {
		return nil, fmt.Errorf("%w: install size %q in header is not a number", ErrCorrupt, header[1])
	}
	// Any hash just means "outdated", but it must survive a round trip.
	if strings.IndexFunc(header[0], func(c rune) bool { return c <= ' ' || c > '~' }) >= 0 {
		return nil, fmt.Errorf("%w: hash %q in header is not printable ASCII", ErrCorrupt, header[0])
	}
	r := &Record{ID: id, Hash: header[0], InstallSize: size}
	if len(header) > 2 {
		r.Depends = header[2:]
	}
	for i, line := range lines[1:] {
		line = strings.TrimRight(line, " \t\r")
		if line == "" {
			continue
		}
		p, err := manifest.CleanRelPath(line)
		if err == nil && p == "" {
			err = fmt.Errorf("%q does not name a file", line)
		}
		if err != nil {
			r.Warnings = append(r.Warnings, fmt.Sprintf("component record %q: skipped line %d: %v", id, i+2, err))
			continue
		}
		r.Files = append(r.Files, p)
	}
	return r, nil
}

// MarshalLegacy encodes r exactly as the C# manager writes it: header
// tokens joined by single spaces, then one backslash path per line, every
// line CRLF-terminated, UTF-8 without BOM.
func (r *Record) MarshalLegacy() []byte {
	var b bytes.Buffer
	b.WriteString(r.Hash + " " + strconv.FormatInt(r.InstallSize, 10))
	for _, d := range r.Depends {
		b.WriteString(" " + d)
	}
	b.WriteString("\r\n")
	for _, f := range r.Files {
		b.WriteString(strings.ReplaceAll(f, "/", `\`) + "\r\n")
	}
	return b.Bytes()
}

// isRecordName mirrors which Components/ entries legacy tools treat as
// records: the C# manager skips names with an extension (Path.HasExtension:
// a "." followed by at least one character), which also hides old
// "<id>.txt" files and components.bak. Dot-files are never records.
func isRecordName(name string) bool {
	i := strings.LastIndexByte(name, '.')
	return name != "" && name[0] != '.' && (i < 0 || i == len(name)-1)
}

// ReadRecords loads every record in <root>/Components. A missing Components
// folder means nothing is installed. One bad record never hides the others:
// each unreadable one, or one with a bad header, is returned as a
// *RecordError in errs (and is not in the map). Records with skipped lines
// are in the map with their Warnings set.
func ReadRecords(root string) (recs map[string]*Record, errs []error) {
	recs = map[string]*Record{}
	entries, err := os.ReadDir(ComponentsDir(root))
	if errors.Is(err, fs.ErrNotExist) {
		return recs, nil
	}
	if err != nil {
		return recs, []error{fmt.Errorf("listing component records: %w", err)}
	}
	for _, e := range entries {
		id := e.Name()
		if !e.Type().IsRegular() || !isRecordName(id) {
			continue
		}
		path := RecordPath(root, id)
		data, err := readLimited(path, 64<<20)
		if errors.Is(err, fs.ErrNotExist) {
			continue // vanished since ReadDir
		}
		var r *Record
		if err == nil {
			r, err = ParseRecord(id, data)
		}
		if err != nil {
			errs = append(errs, &RecordError{ID: id, Path: path, Err: err})
			continue
		}
		recs[id] = r
	}
	return recs, errs
}
