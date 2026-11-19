#!/usr/bin/env bash
# set-version.sh <vX.Y.Z>
#
# Writes a release's version into apps/mirra/nsgo.yaml, from which nsgo build names the app's version and its packages and the website reads the version it shows. The file must hold its version line exactly once, so that a moved or renamed line fails here rather than going unbumped.
set -euo pipefail

tag=$1
version=${tag#v}
if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "not a version: $tag" >&2
  exit 1
fi

file=$(git rev-parse --show-toplevel)/apps/mirra/nsgo.yaml
pattern='^version: [0-9.]+$'
count=$(grep -cE "$pattern" "$file" || true)
if [ "$count" -ne 1 ]; then
  echo "apps/mirra/nsgo.yaml: expected one line matching $pattern, found $count" >&2
  exit 1
fi
sed -E "s/$pattern/version: $version/" "$file" >"$file.tmp"
mv "$file.tmp" "$file"
