package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	mocks "github.com/vektra/mockery/v2/mocks/github.com/vektra/mockery/v2/pkg/fixtures"
)

func TestPanicOnNoReturnValue(t *testing.T) {
	m := mocks.NewPanicOnNoReturnValue(t)
	m.EXPECT().DoSomething()

	defer func() {
		if r := recover(); r != nil {
			assert.NotNil(t, r)
			assert.Equal(t, "Missing Return() function for DoSomething()", r.(string))
		}
	}()

	m.DoSomething()

}
