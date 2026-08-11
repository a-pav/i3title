#!/bin/sh

set -eu

REPO='a-pav/i3title'
API="https://api.github.com/repos/$REPO/releases/latest"

# Default installation path.
PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="${BINDIR:-$PREFIX/bin}"

case "$(uname -m)" in
	x86_64 | amd64 )  ARCH=amd64 ;;
	aarch64 | arm64 ) ARCH=arm64 ;;
	* )
		echo "Unsupported architecture: $(uname -m)" >&2
		exit 1
		;;
esac

echo "Fetching latest release . . ."
release="$(curl -fsS "$API")"

url="$(echo "$release" \
	| grep "browser_download_url" \
	| grep "linux-$ARCH.tar.gz\"" \
	| cut -d '"' -f4
)"
if [ -z "$url" ]; then
	echo 'Release not found!'
	exit 1
fi
filename=${url##*/}

echo "Downloading $filename . . ."

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT HUP INT TERM

archive="$tmpdir/$filename"
curl -fsSL -o "$archive" "$url"

# Check download integrity
digest="$(sha256sum "$archive" | cut -d' ' -f1)"
if ! echo "$release" | grep -q "$digest"; then
	echo "Checksum verification failed." >&2
	exit 1
fi

# Extract archive
mkdir -p "$tmpdir/ext"
tar -C "$tmpdir/ext" -xzf "$archive"

echo "Installing to $BINDIR . . ."
install -d "$BINDIR"
install -m 755 "$tmpdir/ext/i3title" "$BINDIR/i3title"
install -m 755 "$tmpdir/ext/i3toast" "$BINDIR/i3toast"

echo "Done!"
