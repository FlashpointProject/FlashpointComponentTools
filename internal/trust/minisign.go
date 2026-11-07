package trust

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/blake2b"
)

// Minisign formats (https://jedisct1.github.io/minisign/):
//
//	public key:  base64("Ed" || key_id[8] || ed25519_pk[32])
//	signature file, four lines:
//	  untrusted comment: <anything>
//	  base64(sig_alg[2] || key_id[8] || sig[64])
//	  trusted comment: <text>
//	  base64(global_sig[64])     Ed25519 over sig[64] || <text>
//
// Only sig_alg "ED" (Ed25519 over BLAKE2b-512(message), minisign's default)
// is accepted. sig_alg is not signed, so also accepting legacy "Ed" (which
// signs the message itself) would let anyone relabel a valid "ED" signature
// of M as an "Ed" signature of the 64-byte message BLAKE2b(M).
const (
	untrustedPrefix = "untrusted comment: "
	trustedPrefix   = "trusted comment: "
)

// KeyID is a minisign key identifier.
type KeyID [8]byte

// String formats the key ID as minisign prints it, e.g. "6696DDC9C5E924E7".
func (id KeyID) String() string { return fmt.Sprintf("%016X", binary.LittleEndian.Uint64(id[:])) }

// PublicKey is a minisign Ed25519 public key.
type PublicKey struct {
	ID  KeyID
	Key ed25519.PublicKey
}

// ParsePublicKey accepts the base64 key line ("RWT...") or a whole .pub file.
func ParsePublicKey(s string) (PublicKey, error) {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(lines[len(lines)-1]))
	if err != nil || len(lines) > 2 || len(raw) != 2+8+ed25519.PublicKeySize || raw[0] != 'E' || raw[1] != 'd' {
		return PublicKey{}, errors.New("not a minisign Ed25519 public key")
	}
	k := PublicKey{Key: ed25519.PublicKey(raw[10:])}
	copy(k.ID[:], raw[2:10])
	return k, nil
}

// verify checks a minisign signature file (file signature and the global
// signature over the trusted comment) against keys, matched by key ID, and
// returns the trusted comment.
func verify(msg, sigFile []byte, keys []PublicKey) (string, KeyID, error) {
	var id KeyID
	lines := strings.Split(strings.TrimRight(string(sigFile), "\r\n"), "\n")
	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}
	if len(lines) != 4 || !strings.HasPrefix(lines[0], untrustedPrefix) || !strings.HasPrefix(lines[2], trustedPrefix) {
		return "", id, fmt.Errorf("%w: not a minisign signature file", ErrBadSignature)
	}
	sig, err1 := base64.StdEncoding.DecodeString(lines[1])
	global, err2 := base64.StdEncoding.DecodeString(lines[3])
	if err1 != nil || err2 != nil || len(sig) != 2+8+ed25519.SignatureSize || len(global) != ed25519.SignatureSize {
		return "", id, fmt.Errorf("%w: not a minisign signature file", ErrBadSignature)
	}
	if alg := string(sig[:2]); alg != "ED" {
		return "", id, fmt.Errorf("%w: signature algorithm %q is not accepted; sign with minisign 0.11 or later without -l", ErrBadSignature, alg)
	}
	copy(id[:], sig[2:10])
	i := 0
	for i < len(keys) && keys[i].ID != id {
		i++
	}
	if i == len(keys) {
		return "", id, fmt.Errorf("%w: signed with unknown key %s", ErrBadSignature, id)
	}
	comment := strings.TrimPrefix(lines[2], trustedPrefix)
	h := blake2b.Sum512(msg)
	if !ed25519.Verify(keys[i].Key, h[:], sig[10:]) ||
		!ed25519.Verify(keys[i].Key, append(bytes.Clone(sig[10:]), comment...), global) {
		return "", id, fmt.Errorf("%w: signature does not match (key %s)", ErrBadSignature, id)
	}
	return comment, id, nil
}

// Sign produces a minisign signature file ("ED" variant) for msg, for tests
// and publishing tooling. Newlines in trusted become spaces.
func Sign(priv ed25519.PrivateKey, keyID KeyID, msg []byte, trusted string) []byte {
	trusted = strings.NewReplacer("\r", " ", "\n", " ").Replace(trusted)
	h := blake2b.Sum512(msg)
	sig := ed25519.Sign(priv, h[:])
	global := ed25519.Sign(priv, append(bytes.Clone(sig), trusted...))
	line2 := append(append([]byte("ED"), keyID[:]...), sig...)
	enc := base64.StdEncoding.EncodeToString
	return fmt.Appendf(nil, "%ssignature from fpm signing key %s\n%s\n%s%s\n%s\n",
		untrustedPrefix, keyID, enc(line2), trustedPrefix, trusted, enc(global))
}
