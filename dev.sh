#!/usr/bin/env bash
cd "$(dirname "$0")"
wails dev -tags "webkit2_41" "$@"
