#!/usr/bin/env bash
set -eo pipefail
echo "Downloading files..."
requireEnv() {
  test "${!1}" || (echo "server: '$1' not found" >&2 && exit 1)
}
requireEnv STORAGE_BUCKET

gsutil cp -r gs://${STORAGE_BUCKET}/$1 $2

echo "wrote to $2"