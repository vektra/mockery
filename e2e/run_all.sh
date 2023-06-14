#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

for file in $(ls $SCRIPT_DIR/test_*.sh); do
    $file
done
