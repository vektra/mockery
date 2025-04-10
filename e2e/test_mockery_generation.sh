#!/bin/bash

go run github.com/go-task/task/v3/cmd/task mocks
rt=$?
if [ $rt -ne 0 ]; then
    echo "ERROR: non-zero return code from mockery"
    exit 1
fi
echo "SUCCESS: successfully generated mocks defined in .mockery.yaml"