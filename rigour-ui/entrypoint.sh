#!/bin/sh

# Replace placeholder in built files with the actual environment variable value
find /app/.next/ -type f -name '*.js' -exec sed -i "s|NEXT_PUBLIC_API_BASE_URL_PLACEHOLDER|${NEXT_PUBLIC_API_BASE_URL}|g" {} +

# Start the Next.js application
exec "$@"
