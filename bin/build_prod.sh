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
