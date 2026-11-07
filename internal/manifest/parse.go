package manifest

import (
	"bytes"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// MaxManifestSize caps the document (the real ones are ~64 KB); fetch and
// the components.bak reader use the same value.
const MaxManifestSize = 4 << 20

// Options tweak Parse. The zero value is the safe default.
type Options struct {
	// AllowInsecure accepts an http:// base URL, for custom sources the user
	// configured explicitly (and tests).
	AllowInsecure bool
}

// Parse reads and validates a components.xml document.
func Parse(r io.Reader) (*Manifest, error) { return ParseWithOptions(r, Options{}) }

// node is a generic XML element. Decoding into it keeps child elements in
// document order and drops comments, text and processing instructions;
// unknown elements and attributes are then ignored, so the format can grow.
type node struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Children []node     `xml:",any"`
}

func (n *node) attr(name string) (string, bool) {
	for _, a := range n.Attrs {
		if a.Name.Space == "" && a.Name.Local == name {
			return a.Value, true
		}
	}
	return "", false
}

// ParseWithOptions is Parse with explicit options.
//
// Unsafe IDs or paths, duplicate IDs, malformed hashes or numbers and
// dependency cycles reject the whole manifest: a partially accepted one would
// install a different set than the other clients see. Dangling dependencies
// and a "-" inside an id attribute are only warnings (the C# clients
// silently accept both).
func ParseWithOptions(r io.Reader, opt Options) (*Manifest, error) {
	b, err := io.ReadAll(io.LimitReader(r, MaxManifestSize+1))
	if err != nil {
		return nil, fmt.Errorf("read component list: %w", err)
	}
	if len(b) > MaxManifestSize {
		return nil, invalid("larger than %d bytes", MaxManifestSize)
	}
	// encoding/xml rejects non-UTF-8 encoding declarations, but not a BOM.
	var root node
	if err := xml.Unmarshal(bytes.TrimPrefix(b, []byte("\xef\xbb\xbf")), &root); err != nil {
		if errors.Is(err, io.EOF) {
			err = errors.New("empty document")
		}
		return nil, fmt.Errorf("%w: %w", ErrInvalid, err)
	}
	if root.XMLName.Local != "list" {
		return nil, invalid("root element is <%s>, expected <list>", root.XMLName.Local)
	}
	m := &Manifest{Components: map[string]*Component{}, Categories: map[string]*Category{}}
	raw, _ := root.attr("url")
	if m.BaseURL, err = baseURL(raw, opt.AllowInsecure); err != nil {
		return nil, err
	}
	// Only categories are meaningful directly under <list>.
	for i := range root.Children {
		if n := &root.Children[i]; n.XMLName.Local == "category" {
			c, err := m.category(n, nil)
			if err != nil {
				return nil, err
			}
			m.Roots = append(m.Roots, c)
		}
	}
	if err := m.checkDepends(); err != nil {
		return nil, err
	}
	return m, nil
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

// baseURL validates <list url> and makes it end with "/". Archive URLs are
// built by appending "<id>.zip", so a query or fragment would swallow that.
func baseURL(raw string, insecure bool) (string, error) {
	s := strings.TrimSpace(raw)
	u, err := url.Parse(s)
	switch {
	case err != nil || u.Host == "" || u.User != nil || strings.ContainsAny(s, "?#"):
		return "", invalid("list url %q must be an absolute https URL without query, fragment or credentials", raw)
	case u.Scheme == "http" && !insecure:
		return "", invalid("list url %q uses http; only https is allowed", raw)
	case u.Scheme != "https" && u.Scheme != "http":
		return "", invalid("list url %q must be an https URL", raw)
	}
	if !strings.HasSuffix(s, "/") {
		s += "/"
	}
	return s, nil
}

// header validates the attributes shared by categories and components.
func (m *Manifest) header(n *node, parent *Category) (id, local, title, desc string, err error) {
	kind := n.XMLName.Local
	local, _ = n.attr("id")
	id = local
	if parent != nil {
		id = parent.ID + "-" + local
	}
	if !ValidID(id) {
		return "", "", "", "", invalid("%s ID %q is invalid (allowed: lowercase letters, digits and \"_\")", kind, id)
	}
	// Only full IDs must be unique: a clash such as cat + "a-b" vs cat-a + "b"
	// is caught here. The C# clients accept a "-" inside an id attribute, so
	// it is only a warning: `fpm download <category>` matches IDs by prefix
	// and may then pick up a component from another category.
	if m.Components[id] != nil || m.Categories[id] != nil {
		return "", "", "", "", invalid("duplicate ID %q", id)
	}
	if strings.Contains(local, "-") {
		m.Warnings = append(m.Warnings, fmt.Sprintf("%s %q: id %q contains \"-\", which normally separates categories", kind, id, local))
	}
	title, ok1 := n.attr("title")
	desc, ok2 := n.attr("description")
	if !ok1 || !ok2 {
		return "", "", "", "", invalid("%s %q needs a title and a description attribute", kind, id)
	}
	return id, local, cleanText(title, false), cleanText(desc, true), nil
}

func (m *Manifest) category(n *node, parent *Category) (*Category, error) {
	id, local, title, desc, err := m.header(n, parent)
	if err != nil {
		return nil, err
	}
	c := &Category{ID: id, LocalID: local, Title: title, Description: desc, Required: requiredID(id)}
	m.Categories[id] = c
	for i := range n.Children {
		ch := &n.Children[i]
		switch ch.XMLName.Local {
		case "category":
			sub, err := m.category(ch, c)
			if err != nil {
				return nil, err
			}
			c.Items = append(c.Items, Item{Category: sub})
		case "component":
			comp, err := m.component(ch, c)
			if err != nil {
				return nil, err
			}
			c.Items = append(c.Items, Item{Component: comp})
		}
	}
	return c, nil
}

func (m *Manifest) component(n *node, parent *Category) (*Component, error) {
	id, local, title, desc, err := m.header(n, parent)
	if err != nil {
		return nil, err
	}
	c := &Component{ID: id, LocalID: local, Title: title, Description: desc}
	bad := func(format string, args ...any) error {
		return invalid("component %q: %s", id, fmt.Sprintf(format, args...))
	}

	// Required. C# only checks the length; we verify downloads against it.
	hash, _ := n.attr("hash")
	crc, err := strconv.ParseUint(hash, 16, 32)
	if len(hash) != 8 || err != nil {
		return nil, bad("hash %q is not 8 hexadecimal digits", hash)
	}
	c.Hash, c.CRC32 = hash, uint32(crc)

	var modified int64
	for _, f := range []struct {
		name string
		dst  *int64
	}{{"download-size", &c.DownloadSize}, {"install-size", &c.InstallSize}, {"date-modified", &modified}} {
		v, _ := n.attr(f.name)
		if v = strings.TrimSpace(v); v == "" {
			continue // optional, default 0
		}
		if *f.dst, err = strconv.ParseInt(v, 10, 64); err != nil || *f.dst < 0 {
			return nil, bad("%s %q is not a non-negative whole number", f.name, v)
		}
	}
	if modified > 0 {
		c.Modified = time.Unix(modified, 0).UTC()
	}

	p, _ := n.attr("path")
	if c.Path, err = CleanRelPath(p); err != nil {
		return nil, bad("%v", err)
	}

	deps, _ := n.attr("depends")
	for _, dep := range strings.Fields(deps) {
		if !ValidID(dep) {
			return nil, bad("depends on %q, which is not a valid component ID", dep)
		}
		c.Depends = append(c.Depends, dep)
	}

	if v, ok := n.attr("sha256"); ok {
		if c.SHA256, err = hex.DecodeString(v); err != nil || len(c.SHA256) != 32 {
			return nil, bad("sha256 %q is not 64 hexadecimal digits", v)
		}
	}

	DerivePolicy(c)
	m.Components[id] = c
	m.Order = append(m.Order, id)
	return c, nil
}

// checkDepends warns about dangling dependencies and rejects cycles (the C#
// manager overflows its stack on them).
func (m *Manifest) checkDepends() error {
	const active, done = 1, 2
	state := map[string]int{}
	var path []string
	var visit func(id string) error
	visit = func(id string) error {
		switch state[id] {
		case done:
			return nil
		case active:
			cycle := append(slices.Clone(path[slices.Index(path, id):]), id)
			return invalid("dependency cycle: %s", strings.Join(cycle, " -> "))
		}
		state[id] = active
		path = append(path, id)
		for _, dep := range m.Components[id].Depends {
			if m.Components[dep] != nil {
				if err := visit(dep); err != nil {
					return err
				}
			}
		}
		path = path[:len(path)-1]
		state[id] = done
		return nil
	}
	for _, id := range m.Order {
		for _, dep := range m.Components[id].Depends {
			if m.Components[dep] == nil {
				m.Warnings = append(m.Warnings, fmt.Sprintf("component %q depends on %q, which does not exist (ignored)", id, dep))
			}
		}
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}

// cleanText turns control characters into spaces. Titles are printed by
// `fpm list`, which the launcher parses line by line, so they must never
// contain a newline; descriptions may.
func cleanText(s string, keepNewlines bool) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && !(keepNewlines && r == '\n') {
			return ' '
		}
		return r
	}, s)
}
