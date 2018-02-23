#!/bin/bash
set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."
source .envrc

go get -u golang.org/x/vgo

vgo install github.com/onsi/ginkgo/ginkgo
vgo install github.com/cloudfoundry/libbuildpack/packager/buildpack-packager
