#!/bin/bash

cd "$(dirname "$0")/.." ||  exit 10
source ./bin/lib/strict_mode.sh

hook_worker -config=config_prod.yaml
