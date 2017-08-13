#!/bin/bash

source ./bin/lib/strict_mode.sh

task::server() {
  ./bin/prod_server.sh
}

task::worker() {
  ./bin/prod_worker.sh
}

task::build() {
  ./bin/build_prod.sh
}

main() {
  cd $(dirname "$0")/.. || exit 10
  task_name="$1"
  if type "task::${task_name}" &>/dev/null; then
    shift
    eval "task::${task_name}" "$@"
  else
    usage "$@"
  fi
}

main "$@"

