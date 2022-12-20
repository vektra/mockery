#!/usr/bin/env bash

set -euo pipefail

ROOT=$(cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)
cd $ROOT

reset() {
  rm -rf mocks
}

verify() {
  if [ ! -d "mocks" ]; then \
    echo "No Mock Dir Created"; \
    exit 1; \
  fi
  if [ ! -f "mocks/AsyncProducer.go" ]; then \
    echo "AsyncProducer.go not created"; \
    echo 1; \
  fi
}

trap reset exit

reset
${GOPATH}/bin/mockery -all -recursive -cpuprofile="mockery.prof" -dir="mockery/fixtures"
verify

reset
${GOPATH}/bin/mockery -all -recursive -cpuprofile="mockery.prof" -srcpkg github.com/vektra/mockery/pkg/fixtures
verify


reset
docker run -v $(pwd):/src -w /src --user=$(id -u):$(id -g) vektra/mockery -all -recursive -cpuprofile="mockery.prof" -dir="mockery/fixtures"
verify

reset
docker run -v $(pwd):/src -w /src --user=$(id -u):$(id -g) vektra/mockery -all -recursive -cpuprofile="mockery.prof" -srcpkg github.com/vektra/mockery/pkg/fixtures
verify
