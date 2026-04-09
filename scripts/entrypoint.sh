#!/bin/sh
set -e

if [ -n "$BUCKET_NAME" ]; then
  # Restore from replica if no local DB exists yet (new machine or wiped volume)
  litestream restore -if-replica-exists -config /app/litestream.yml /data/openpadel.db
  # Run app under litestream — replicates continuously
  exec litestream replicate -exec "/app/openpadel" -config /app/litestream.yml
else
  exec /app/openpadel
fi
