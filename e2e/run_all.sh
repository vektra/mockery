#!/bin/bash
set +e
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

for test in $(ls -d $SCRIPT_DIR/test_*); do
    file="$test"
    if [ -d "$test" ]; then
        file="$test/run.sh"
    fi
    echo "=========="
    echo "RUNNING $file"
    echo "=========="
    $file
done