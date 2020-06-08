#!/bin/sh

echo "Starting Userme..."
set -x
userme-demo-api \
     --loglevel=$LOG_LEVEL \
     --cors-allowed-origins=$CORS_ALLOWED_ORIGINS \
     --jwt-signing-key-file=$JWT_SIGNING_KEY_FILE \
     --jwt-signing-method=$JWT_SIGNING_METHOD \
     --base-url=$BASE_SERVER_URL_FOR_LOCATIONS
     

