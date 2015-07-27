package main

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestFilename(t *testing.T) {
	var actual string

	actual = filename("Writer", false)
	assert.Equal(t, "writer.go", actual)

	actual = filename("RoundTripper", false)
	assert.Equal(t, "roundtripper.go", actual)

	actual = filename("Writer", true)
	assert.Equal(t, "mock_writer_test.go", actual)

	actual = filename("RoundTripper", true)
	assert.Equal(t, "mock_roundtripper_test.go", actual)
}
