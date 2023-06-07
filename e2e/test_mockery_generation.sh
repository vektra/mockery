#!/bin/bash

echo "=========="
echo "RUNNING $0"
echo "=========="

go run .
rt=$?
if [ $rt -ne 0 ]; then
    echo "ERROR: non-zero return code from mockery"
    exit 1
fi
echo "SUCCESS: successfully generated mocks defined in .mockery.yaml"