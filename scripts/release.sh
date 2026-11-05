#!/usr/bin/env bash
# Build, sign, notarize and upload a JustLearn release to GitHub.
# The update signing key stays outside the repository.
set -euo pipefail
cd "$(dirname "$0")/.."

key_file="${MYGO_UPDATER_PRIVATE_KEY_FILE:-$HOME/Library/Application Support/mygo/update-keys/justlearn/mygo-update.key}"
if [[ -z "${MYGO_UPDATER_PRIVATE_KEY:-}" ]]; then
  if [[ ! -f "$key_file" ]]; then
    echo "missing update signing key: $key_file" >&2
    echo "run 'go tool mygo keygen -o ...' from desktop/ first" >&2
    exit 1
  fi
  MYGO_UPDATER_PRIVATE_KEY="$(cat "$key_file")"
  export MYGO_UPDATER_PRIVATE_KEY
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "working tree is dirty; commit the release first" >&2
  exit 1
fi

version="$(python3 -c 'import json; print(json.load(open("desktop/mygo.json"))["version"])')"
echo "Building JustLearn $version (darwin/universal)"

cd desktop
go tool mygo build -platform darwin/universal -upload "$@"

echo
echo "Draft release: v$version"
echo "Review the draft, then publish with:"
echo "  gh release edit v$version --draft=false"

if [[ "${PUBLISH:-0}" == "1" ]]; then
  gh release edit "v$version" --draft=false
  echo "Published v$version"
fi
