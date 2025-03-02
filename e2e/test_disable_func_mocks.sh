#!/bin/bash
go run github.com/go-task/task/v3/cmd/task mocks.remove || exit 1
go run github.com/go-task/task/v3/cmd/task mocks.generate || exit 1

export MOCKERY_CONFIG="e2e/.mockery-disable-func-mock.yaml"
export MOCKERY_LOG_LEVEL="error"

MOCKERY_DISABLE_FUNC_MOCKS="false" go run github.com/go-task/task/v3/cmd/task mocks.generate

if [ -f "./mocks/github.com/vektra/mockery/v2/pkg/fixtures/mock_SendFunc.go" ]; then
    echo "file exists as expected"
else
    echo "file doesn't exist when we expected it to exist"
    exit 1
fi

go run github.com/go-task/task/v3/cmd/task mocks.remove
MOCKERY_DISABLE_FUNC_MOCKS="true" go run github.com/go-task/task/v3/cmd/task mocks.generate
if [ -f "./mocks/github.com/vektra/mockery/v2/pkg/fixtures/mock_SendFunc.go" ]; then
    echo "SendFunc mock exists when we expected it to not be generated."
    exit 1
else
    echo "SendFunc mock doesn't exist as expected"
fi