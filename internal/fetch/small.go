package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// FetchManifest GETs a manifest: one request, no retries, time budget from
// ctx. Nothing is stored: the caller verifies and parses the body first.
// The manifest is small and fetched once per start, so there is no
// conditional request or local cache (components.bak is the offline copy).
func (c *Client) FetchManifest(ctx context.Context, u string) ([]byte, error) {
	body, err := c.FetchSmall(ctx, u, MaxManifestBytes)
	if err != nil {
		return nil, fmt.Errorf("download component list: %w", err)
	}
	return body, nil
}

// FetchSmall GETs a small file such as a ".minisig" signature. A 404/410
// returns an error matching ErrNotFound, so the caller can tell "unsigned"
// from "server down".
func (c *Client) FetchSmall(ctx context.Context, u string, maxBytes int64) ([]byte, error) {
	resp, err := c.get(ctx, u, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.CopyN(io.Discard, resp.Body, 4<<10) // lets the connection be reused
		return nil, newStatusError(resp, u)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", u, c.netErr(ctx, err))
	}
	if int64(len(b)) > maxBytes {
		return nil, fmt.Errorf("%s: response is larger than %d bytes", u, maxBytes)
	}
	return b, nil
}
