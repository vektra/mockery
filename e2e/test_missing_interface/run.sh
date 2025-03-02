#!/bin/bash
set +e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
CONFIG=$SCRIPT_DIR/.mockery.yml
export MOCKERY_CONFIG=$CONFIG

go run github.com/go-task/task/v3/cmd/task mocks.generate

RT=$?
if [ $RT -eq 0 ]; then
    echo "ERROR: Expected mockery to fail."
    exit 1
fi
echo "SUCCESS: Mockery returned non-zero return code as expected."
