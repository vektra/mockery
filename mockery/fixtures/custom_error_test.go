package test_test

import (
	"testing"

	test "github.com/vektra/mockery/mockery/fixtures"
)

func TestCustomError(t *testing.T) {
	msg := "answer known; question unknowable"
	err := test.NewErr(msg, 42)

	if err.Error() != msg {
		t.Error(err.Error(), "!=", msg)
	}
}
