#!/bin/sh
# Builds fpm and FlashpointManager for the release targets into
# dist/<goos>-<goarch>/, plus dist/windows-amd64/core-manager.zip and
# dist/windows-amd64/FlashpointInstaller.exe (Windows only; it is a separate
# download, not part of core-manager.zip).
#
#   scripts/build.sh [version]      (default: git describe, else "dev")
#
# windows/amd64 is the shipping Windows build (DESIGN.md §2): 32-bit Windows
# is no longer supported, in line with the launcher.
#
# The GUI (Gio) builds for Windows without cgo, but needs cgo and the native
# SDK on macOS and Linux, so those manager binaries are only built on a
# matching host; fpm cross-builds everywhere.
#
# Windows icon + manifest: cmd/{manager,fpm,installer}/rsrc_windows_*.syso
# are committed, generated from cmd/*/winres/ (winres.json and icon.ico).
# After changing those, regenerate them with (a build tool only, not a
# go.mod dependency):
#
#   (cd cmd/manager   && go run github.com/tc-hib/go-winres@v0.3.3 make --arch amd64)
#   (cd cmd/fpm       && go run github.com/tc-hib/go-winres@v0.3.3 make --arch amd64)
#   (cd cmd/installer && go run github.com/tc-hib/go-winres@v0.3.3 make --arch amd64)
set -eu
cd "$(dirname "$0")/.."

version=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}
targets="windows/amd64 linux/amd64 darwin/arm64"
host="$(go env GOHOSTOS)/$(go env GOHOSTARCH)"

build() { # build <goos> <goarch> <cgo> <out> <pkg> [extra ldflags]
	echo "building $4 ($version)"
	CGO_ENABLED=$3 GOOS=$1 GOARCH=$2 go build -trimpath \
		-ldflags "-s -w -X main.version=$version ${6:-}" -o "$4" "$5"
}

for t in $targets; do
	goos=${t%/*}
	goarch=${t#*/}
	dir="dist/$goos-$goarch"
	if [ "$goos" = windows ]; then
		build "$goos" "$goarch" 0 "$dir/fpm.exe" ./cmd/fpm
		build "$goos" "$goarch" 0 "$dir/FlashpointManager.exe" ./cmd/manager -H=windowsgui
		build "$goos" "$goarch" 0 "$dir/FlashpointInstaller.exe" ./cmd/installer -H=windowsgui
		continue
	fi
	build "$goos" "$goarch" 0 "$dir/fpm" ./cmd/fpm
	if [ "$t" = "$host" ]; then
		build "$goos" "$goarch" 1 "$dir/FlashpointManager" ./cmd/manager
	else
		echo "skipping $dir/FlashpointManager: the GUI needs cgo, build it on $t"
	fi
done

# The core-manager component (path "Manager") ships both executables at the
# zip root (DESIGN.md §1.1).
zipdir=dist/windows-amd64
rm -f "$zipdir/core-manager.zip"
if command -v zip >/dev/null; then
	(cd "$zipdir" && zip -q -X core-manager.zip FlashpointManager.exe fpm.exe)
else # Git Bash has no zip; Windows' own bsdtar (not Git's GNU tar) writes one
	bsdtar="$(cygpath -u "$SYSTEMROOT")/System32/tar.exe"
	(cd "$zipdir" && "$bsdtar" -a -cf core-manager.zip FlashpointManager.exe fpm.exe)
fi
z=$zipdir/core-manager.zip
size=$(wc -c <"$z" | tr -d ' ')
install=$(cat "$zipdir/FlashpointManager.exe" "$zipdir/fpm.exe" | wc -c | tr -d ' ')
# CRC32 of the whole zip (the manifest's hash): the last 8 bytes of a gzip
# stream are the input's CRC32 and size, little-endian (as is the host).
crc=$(gzip -c <"$z" | tail -c 8 | od -An -tx4 -N4 | tr -d ' \n')
if command -v sha256sum >/dev/null; then
	sha=$(sha256sum <"$z" | cut -d' ' -f1)
else
	sha=$(shasum -a 256 <"$z" | cut -d' ' -f1)
fi
echo
echo "$z"
echo "  hash (crc32)   $crc"
echo "  download-size  $size"
echo "  install-size   $install"
echo "  sha256         $sha"
