#!/usr/bin/env bash
set -euo pipefail

cleanup() {
  trap - EXIT INT TERM
  kill 0
}
trap cleanup EXIT INT TERM

( cd backend && air ) &
( cd frontend && pnpm dev ) &

wait
