// Package trust verifies manifest signatures (minisign, Ed25519) against
// keys compiled in per channel. While a channel has no key, verification is
// inactive and every manifest is accepted as Unsigned. Once it has keys, its
// manifests must carry a valid signature whose trusted comment names the
// channel. Rollback and expiry protection are left for when a real signing
// key and key policy exist.
package trust

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Result of checking a manifest.
type Result int

const (
	Unsigned Result = iota // the channel has no keys: nothing was verified
	Valid                  // signature valid under an embedded key
)

func (r Result) String() string {
	if r == Valid {
		return "signed"
	}
	return "unsigned"
}

// ErrBadSignature is wrapped by every Check error: a missing or malformed
// signature, an unknown key, failed verification, or a signature made for
// another channel. The message is user-facing.
var ErrBadSignature = errors.New("manifest signature is invalid (the download may be corrupted or tampered with)")

// embeddedKeys maps channel -> accepted keys (current and next). The
// development CI key must never be listed for stable. Add keys from
// `minisign -G` through a helper that panics on a typo, so a bad key fails
// the tests instead of disabling verification. To rotate, ship the new key
// in the second slot, start signing with it, drop the old one later.
var embeddedKeys = map[string][]PublicKey{}

// Keys returns the compiled-in keys for channel (none yet for any channel).
func Keys(channel string) []PublicKey { return embeddedKeys[channel] }

// FormatComment builds the trusted comment a manifest signature for channel
// must carry. Other space-separated fields may follow.
func FormatComment(channel string) string { return "fpm-manifest channel=" + channel }

// Check verifies manifest against its .minisig body sig (nil if the server
// has none). With no keys it returns Unsigned and no error. With keys it
// fails closed: the signature must exist, verify under one of the keys, and
// name channel in its trusted comment.
func Check(keys []PublicKey, channel string, manifest, sig []byte) (Result, error) {
	if len(keys) == 0 {
		return Unsigned, nil
	}
	if len(sig) == 0 {
		return Unsigned, fmt.Errorf("%w: the %q component list is not signed", ErrBadSignature, channel)
	}
	comment, _, err := verify(manifest, sig, keys)
	if err != nil {
		return Unsigned, err
	}
	if !slices.Contains(strings.Fields(comment), "channel="+channel) {
		return Unsigned, fmt.Errorf("%w: it was not signed for channel %q (trusted comment %q)", ErrBadSignature, channel, comment)
	}
	return Valid, nil
}
