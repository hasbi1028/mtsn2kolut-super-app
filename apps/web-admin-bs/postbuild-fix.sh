#!/bin/bash
# Post-build fix: patch asset_dir path in the SvelteKit adapter-node build output.
# The chunk file lives in build/server/chunks/ but the client assets are in build/client/.
# The built-in path resolves to build/server/chunks/client which doesn't exist.
# This script patches it to ../../client relative to the chunk directory.

set -euo pipefail

CHUNK_DIR="build/server/chunks"
CLIENT_DIR="build/client"

if [ ! -d "$CLIENT_DIR" ]; then
  echo "[postbuild-fix] No build/client/ directory found, skipping."
  exit 0
fi

# Find the main chunk file that contains asset_dir
CHUNK_FILE=$(grep -rl 'const asset_dir' "$CHUNK_DIR"/*.js 2>/dev/null | head -1)

if [ -z "$CHUNK_FILE" ]; then
  echo "[postbuild-fix] No chunk with asset_dir found, skipping."
  exit 0
fi

# Patch asset_dir to use ../../client instead of ./client
if grep -q 'dir}/client' "$CHUNK_FILE"; then
  sed -i 's|`\\${dir}/client|`${dir}/../../client|g' "$CHUNK_FILE"
  echo "[postbuild-fix] Patched asset_dir in $CHUNK_FILE"
fi

# Patch serve() call that uses dir/client for static file serving
if grep -q "path.join(dir, 'client')" "$CHUNK_FILE"; then
  sed -i "s|path.join(dir, 'client')|path.join(dir, '../../client')|g" "$CHUNK_FILE"
  echo "[postbuild-fix] Patched serve() client path in $CHUNK_FILE"
fi

# Patch prerendered path too
if grep -q "path.join(dir, 'prerendered')" "$CHUNK_FILE"; then
  sed -i "s|path.join(dir, 'prerendered')|path.join(dir, '../../prerendered')|g" "$CHUNK_FILE"
  echo "[postbuild-fix] Patched serve() prerendered path in $CHUNK_FILE"
fi

echo "[postbuild-fix] Done."
