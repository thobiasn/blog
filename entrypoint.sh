#!/bin/sh
set -e

if [ -n "$LITESTREAM_REPLICA_BUCKET" ]; then
    litestream restore -if-replica-exists -o /app/blog.db /app/blog.db
    exec litestream replicate -exec "./blog serve"
else
    exec ./blog serve
fi
