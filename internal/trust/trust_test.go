package trust

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/blake2b"
)

type testKey struct {
	priv ed25519.PrivateKey
	pub  PublicKey
}

// newKey derives a deterministic test key from seed.
func newKey(seed byte) testKey {
	priv := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{seed}, ed25519.SeedSize))
	var id KeyID
	for i := range id {
		id[i] = seed + byte(i)
	}
	return testKey{priv: priv, pub: PublicKey{ID: id, Key: priv.Public().(ed25519.PublicKey)}}
}

func (k testKey) sign(msg []byte, channel string) []byte {
	return Sign(k.priv, k.pub.ID, msg, FormatComment(channel)+" seq=5")
}

func pubLine(k PublicKey) string {
	return base64.StdEncoding.EncodeToString(append(append([]byte("Ed"), k.ID[:]...), k.Key...))
}

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestRealMinisignVectors uses files produced by minisign 0.12 (testdata/).
// The legacy (-l, "Ed") vector must be rejected: sig_alg is not signed, so
// accepting "Ed" would let anyone relabel a valid "ED" signature of M and
// present the 64-byte BLAKE2b(M) as a verified manifest.
func TestRealMinisignVectors(t *testing.T) {
	pub, err := ParsePublicKey(string(readTestdata(t, "test.pub")))
	if err != nil {
		t.Fatal(err)
	}
	if got := pub.ID.String(); got != "6696DDC9C5E924E7" { // as printed in test.pub
		t.Errorf("key ID = %s", got)
	}
	keys := []PublicKey{pub}

	msg, sig := readTestdata(t, "manifest.xml"), readTestdata(t, "manifest.xml.minisig")
	if res, err := Check(keys, "stable", msg, sig); err != nil || res != Valid {
		t.Fatalf("real vector: %v %v", res, err)
	}
	if _, err := Check(keys, "development", msg, sig); !errors.Is(err, ErrBadSignature) {
		t.Errorf("signed for stable, accepted for development: %v", err)
	}
	tampered := bytes.Clone(msg)
	tampered[len(tampered)/2] ^= 1
	if _, err := Check(keys, "stable", tampered, sig); !errors.Is(err, ErrBadSignature) {
		t.Errorf("tampered: %v", err)
	}

	msg, sig = readTestdata(t, "manifest-legacy.xml"), readTestdata(t, "manifest-legacy.xml.minisig")
	if _, err := Check(keys, "development", msg, sig); !errors.Is(err, ErrBadSignature) || !strings.Contains(err.Error(), `"Ed"`) {
		t.Errorf("legacy vector: %v", err)
	}

	// The relabelling attack on our own signature.
	k := newKey(1)
	m := []byte("<list/>")
	lines := strings.Split(string(k.sign(m, "stable")), "\n")
	raw, _ := base64.StdEncoding.DecodeString(lines[1])
	raw[1] = 'd' // "ED" -> "Ed"
	lines[1] = base64.StdEncoding.EncodeToString(raw)
	h := blake2b.Sum512(m)
	if res, err := Check([]PublicKey{k.pub}, "stable", h[:], []byte(strings.Join(lines, "\n"))); res == Valid || !errors.Is(err, ErrBadSignature) {
		t.Errorf("relabelled: %v %v", res, err)
	}
}

// TestSignInteropWithMinisign checks Sign's output with the real minisign
// tool (skipped if it is not installed).
func TestSignInteropWithMinisign(t *testing.T) {
	bin, err := exec.LookPath("minisign")
	if err != nil {
		t.Skip("minisign not installed")
	}
	k := newKey(3)
	msg := filepath.Join(t.TempDir(), "m.xml")
	os.WriteFile(msg, []byte("<list/>\n"), 0o644)
	os.WriteFile(msg+".minisig", k.sign([]byte("<list/>\n"), "stable"), 0o644)
	out, err := exec.Command(bin, "-V", "-P", pubLine(k.pub), "-m", msg).CombinedOutput()
	if err != nil || !strings.Contains(string(out), FormatComment("stable")) {
		t.Fatalf("minisign -V: %v\n%s", err, out)
	}
}

func TestVerify(t *testing.T) {
	stable, stableNext, dev := newKey(1), newKey(2), newKey(9)
	keys := map[string][]PublicKey{"stable": {stable.pub, stableNext.pub}, "development": {dev.pub}}
	msg := []byte(`<list url="https://example.invalid/"></list>`)
	good := stable.sign(msg, "stable")
	impostor := newKey(4) // same key ID as stable, different key material
	impostor.pub.ID = stable.pub.ID
	lines := strings.Split(string(good), "\n")
	raw, _ := base64.StdEncoding.DecodeString(lines[1])
	raw[20] ^= 0x80
	tamperedSig := []byte(strings.Join([]string{lines[0], base64.StdEncoding.EncodeToString(raw), lines[2], lines[3]}, "\n"))

	tests := []struct {
		name    string
		msg     []byte
		sig     []byte
		channel string
		ok      bool
	}{
		{"valid", msg, good, "stable", true},
		{"valid with rotation key", msg, stableNext.sign(msg, "stable"), "stable", true},
		{"valid CRLF", msg, bytes.ReplaceAll(good, []byte("\n"), []byte("\r\n")), "stable", true},
		{"valid without final newline", msg, bytes.TrimSuffix(good, []byte("\n")), "stable", true},
		{"tampered manifest", append(bytes.Clone(msg), ' '), good, "stable", false},
		{"tampered trusted comment", msg, bytes.Replace(good, []byte("seq=5"), []byte("seq=6"), 1), "stable", false},
		{"tampered signature", msg, tamperedSig, "stable", false},
		{"unknown key id", msg, newKey(7).sign(msg, "stable"), "stable", false},
		{"same key id, other key", msg, impostor.sign(msg, "stable"), "stable", false},
		{"dev key not valid for stable", msg, dev.sign(msg, "stable"), "stable", false},
		{"stable key not valid for dev", msg, stable.sign(msg, "development"), "development", false},
		{"comment names other channel", msg, stable.sign(msg, "development"), "stable", false},
		{"channel only as prefix", msg, stable.sign(msg, "stablex"), "stable", false},
		{"comment without channel", msg, Sign(stable.priv, stable.pub.ID, msg, "timestamp:1 file:x"), "stable", false},
		{"garbage", msg, []byte("hello"), "stable", false},
		{"three lines", msg, []byte(strings.Join(lines[:3], "\n")), "stable", false},
		{"lines swapped", msg, []byte(strings.Join([]string{lines[2], lines[1], lines[0], lines[3]}, "\n")), "stable", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Check(keys[tt.channel], tt.channel, tt.msg, tt.sig)
			if tt.ok != (err == nil && res == Valid) || !tt.ok && !errors.Is(err, ErrBadSignature) {
				t.Fatalf("got %v, %v", res, err)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	k := newKey(1)
	keys := []PublicKey{k.pub}
	msg := []byte("<list/>")
	for _, tt := range []struct {
		name string
		keys []PublicKey
		sig  []byte
		want Result
		ok   bool
	}{
		{"no keys: inactive", nil, nil, Unsigned, true},
		{"no keys, signature ignored", nil, []byte("garbage"), Unsigned, true},
		{"keys, valid", keys, k.sign(msg, "stable"), Valid, true},
		{"keys, no signature", keys, nil, Unsigned, false},
		{"keys, empty signature", keys, []byte{}, Unsigned, false},
	} {
		res, err := Check(tt.keys, "stable", msg, tt.sig)
		if res != tt.want || (err == nil) != tt.ok || err != nil && !errors.Is(err, ErrBadSignature) {
			t.Errorf("%s: %v, %v", tt.name, res, err)
		}
	}
}

func TestParsePublicKey(t *testing.T) {
	k := newKey(5).pub
	good := pubLine(k)
	for in, ok := range map[string]bool{
		good:               true,
		"  " + good + "\n": true,
		"untrusted comment: minisign public key X\n" + good + "\n":     true,
		"untrusted comment: minisign public key X\r\n" + good + "\r\n": true,
		"":            false,
		"not base64!": false,
		base64.StdEncoding.EncodeToString([]byte("Ed short")):                        false,
		base64.StdEncoding.EncodeToString(append([]byte("XX"), make([]byte, 40)...)): false,
		good + "\n" + good + "\n" + good:                                             false,
	} {
		got, err := ParsePublicKey(in)
		if (err == nil) != ok || ok && (got.ID != k.ID || !got.Key.Equal(k.Key)) {
			t.Errorf("%.40q: %+v %v", in, got, err)
		}
	}
}

func TestEmbeddedKeys(t *testing.T) {
	for ch, ks := range embeddedKeys {
		if len(ks) == 0 || len(ks) > 2 {
			t.Errorf("channel %q: %d keys, want 1 or 2", ch, len(ks))
		}
	}
}

func FuzzVerify(f *testing.F) {
	k := newKey(1)
	msg := []byte("<list/>")
	f.Add(msg, k.sign(msg, "stable"))
	f.Add(msg, []byte("untrusted comment: x\nRUQ=\ntrusted comment: y\nAAAA\n"))
	f.Fuzz(func(t *testing.T, m, sig []byte) {
		res, err := Check([]PublicKey{k.pub}, "stable", m, sig)
		if err == nil && (res != Valid || !bytes.Equal(m, msg)) {
			t.Fatalf("forged verification: %q", m)
		}
		if err != nil && !errors.Is(err, ErrBadSignature) {
			t.Fatalf("unexpected error class: %v", err)
		}
	})
}
