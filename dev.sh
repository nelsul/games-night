#!/bin/sh
# Runs the API against the local Postgres and SeaweedFS S3 (docker compose -f docker-compose.dev.yml up -d).
# Login codes are printed in this terminal because SMTP is not configured.
cd "$(dirname "$0")/api"
export DATABASE_URL=postgres://gamesnight:gamesnight@127.0.0.1:5435/gamesnight
export S3_ENDPOINT=127.0.0.1:8334 S3_ACCESS_KEY=gamesnight S3_SECRET_KEY=gamesnight-dev-secret
export S3_BUCKET=games-night S3_USE_SSL=false S3_REGION=us-east-1 S3_CREATE_BUCKET=true
export ADDR=127.0.0.1:8090 PUBLIC_URL=http://localhost:5175
exec go run .
