#!/bin/bash

cd "$(dirname "$0")/.." ||  exit 10
source ./bin/lib/strict_mode.sh

hook_server -config=config_prod.yaml
hook_worker -config=config_prod.yaml
