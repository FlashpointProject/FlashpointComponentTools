package manifest

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// Full IDs are "-"-joined segments of [a-z0-9_]. "." is not allowed: the C#
// manager treats Components/ files with an extension as "not a record"
// (old installs still have "<id>.txt" files there).
var idRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_]*(-[a-z0-9][a-z0-9_]*)*$`)

// ValidID reports whether s is a valid full component or category ID, i.e.
// safe as a file name in Components/ and in URLs. The length cap keeps
// Components/<id> and <id>.zip under every filesystem's 255-byte limit.
func ValidID(s string) bool { return len(s) <= 200 && idRe.MatchString(s) && !reservedName(s) }

// ErrUnsafePath is wrapped by CleanRelPath errors.
var ErrUnsafePath = errors.New("unsafe path")

// CleanRelPath validates a Flashpoint-relative path from untrusted input
// (manifest path, record lines, zip entry names) and returns it
// slash-separated without empty or "." segments or surrounding slashes; ""
// is the Flashpoint root. Backslashes count as separators.
//
// The path must be safe under both Windows and Unix rules, so a manifest
// can't behave differently per OS: no absolute or UNC paths, no ".."
// segments, no ":" (drive letters, alternate data streams), control
// characters or <>"|?*, no segment ending in "." or " " (Windows strips them,
// so "Data." would alias "Data"), and no reserved device names.
//
// This is the only path validator: txn.CheckPath and state delegate to it.
func CleanRelPath(p string) (string, error) {
	fail := func(why string) (string, error) { return "", fmt.Errorf("%w: %q %s", ErrUnsafePath, p, why) }
	if !utf8.ValidString(p) {
		return fail("is not valid UTF-8")
	}
	s := strings.ReplaceAll(p, `\`, "/")
	if strings.HasPrefix(s, "/") {
		return fail("is absolute")
	}
	var out []string
	for _, seg := range strings.Split(s, "/") {
		switch {
		case seg == "" || seg == ".":
			continue
		case seg == "..":
			return fail(`contains ".."`)
		case len(seg) > 255:
			return fail("has a segment longer than 255 bytes")
		case strings.ContainsFunc(seg, func(r rune) bool { return r < 0x20 || r == 0x7f }):
			return fail("contains a control character")
		case strings.ContainsAny(seg, `<>:"|?*`):
			return fail(`contains one of <>:"|?*, which Windows does not allow`)
		case strings.HasSuffix(seg, ".") || strings.HasSuffix(seg, " "):
			return fail("has a segment ending with a dot or space")
		case reservedName(seg):
			return fail("contains a reserved Windows device name")
		}
		out = append(out, seg)
	}
	return strings.Join(out, "/"), nil
}

// Windows device names.
var reservedNames = strings.Fields(`CON PRN AUX NUL CONIN$ CONOUT$ CLOCK$
	COM0 COM1 COM2 COM3 COM4 COM5 COM6 COM7 COM8 COM9
	LPT0 LPT1 LPT2 LPT3 LPT4 LPT5 LPT6 LPT7 LPT8 LPT9`)

// reservedName reports whether Windows treats a file name as a device. It
// ignores everything from the first "." and spaces before it, so "nul.txt"
// and "Nul .tar.gz" are the NUL device too.
func reservedName(name string) bool {
	stem, _, _ := strings.Cut(name, ".")
	return slices.Contains(reservedNames, strings.ToUpper(strings.TrimRight(stem, " ")))
}
