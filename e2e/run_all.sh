#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

for file in $(ls $SCRIPT_DIR/test_*.sh); do
    echo "=========="
    echo "RUNNING $file"
    echo "=========="
    go run github.com/go-task/task/v3/cmd/task mocks.remove || exit 1
    go run github.com/go-task/task/v3/cmd/task mocks.generate || exit 1
    $file
done
