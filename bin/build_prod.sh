#!/bin/bash
#Get/Vendor dependencies for prodution

cd "$(dirname "$0")/.." || exit 10
source ./bin/lib/strict_mode.sh

# clean dependencies
rm -rf vendor
go clean ./...

go get ./...
go get -u github.com/kardianos/govendor
govendor init
govendor fetch ./...

# Add heroku details for vendor.json
cat ./vendor/vendor.json | jq --argjson \
  heroku '{"install": ["./cmd/..."], "goVersion": "go1.8.3"}' '. + {heroku: $heroku}' > \
  vendor_temp.json && mv vendor_temp.json ./vendor/vendor.json
