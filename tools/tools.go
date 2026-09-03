//go:build tools

package main

import (
	_ "github.com/go-task/task/v3/cmd/task"
	_ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
	_ "gotest.tools/gotestsum"
)
