#!/bin/bash

cd "$(dirname "$0")/.." ||  exit 10
source ./bin/lib/strict_mode.sh

./cmd/hook_server -config=config_prod.yaml
