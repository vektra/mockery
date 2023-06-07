#!/bin/bash
# This tests https://github.com/vektra/mockery/issues/632, where
# mockery was generating mocks of its own auto-generated code.

echo "=========="
echo "RUNNING $0"
echo "=========="

# New mocks may legimitately be created, so we run mockery once first
go run .
num_files_before=$(find . -type f | wc -l)
go run .
num_files_after=$(find . -type f | wc -l)

if [ $num_files_before -ne $num_files_after ]; then
    echo "ERROR: detected increased file count over multiple mockery runs."
    echo "before: $num_files_before. after: $num_files_after"
    exit 1
fi
echo "SUCCESS: identical number of files over multiple mockery runs"

