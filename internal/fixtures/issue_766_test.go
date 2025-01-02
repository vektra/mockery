package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Asserts it implements the interface
var _ Issue766 = new(MockIssue766)

func TestIssue766(t *testing.T) {
	fetchFunc := func(i ...int) ([]int, error) {
		ret := make([]int, 0, len(i))
		for idx := 0; idx < len(i); idx++ {
			ret[idx] = i[idx] + 1
		}
		return ret, nil
	}

	expected := []int{1, 2, 3}
	mockFetchData := NewMockIssue766(t)
	mockFetchData.
		EXPECT().
		FetchData(mock.AnythingOfType("func(...int) ([]int, error)")).
		Return([]int{1, 2, 3}, nil)

	actual, err := mockFetchData.FetchData(fetchFunc)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
