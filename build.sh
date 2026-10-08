#!/usr/bin/env bash
cd "$(dirname "$0")"
wails build -tags "webkit2_41" "$@"
