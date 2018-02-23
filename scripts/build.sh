#!/usr/bin/env bash
set -ex

ROOTDIR="$( dirname "$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" )"
BINDIR=$ROOTDIR/bin

export GOPATH=$ROOTDIR
export GOOS=linux

vgo build -ldflags="-s -w" -o $BINDIR/varify nginx/varify

vgo build -ldflags="-s -w" -o $BINDIR/supply nginx/supply/cli
vgo build -ldflags="-s -w" -o $BINDIR/finalize nginx/finalize/cli
