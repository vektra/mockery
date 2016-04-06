package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnderscoreCaseName(t *testing.T) {
	assert.Equal(t, "notify_event", underscoreCaseName("NotifyEvent"))
	assert.Equal(t, "repository", underscoreCaseName("Repository"))
	assert.Equal(t, "http_server", underscoreCaseName("HTTPServer"))
	assert.Equal(t, "awesome_http_server", underscoreCaseName("AwesomeHTTPServer"))
	assert.Equal(t, "csv", underscoreCaseName("CSV"))
	assert.Equal(t, "position0_size", underscoreCaseName("Position0Size"))
}

func TestFilenameBare(t *testing.T) {
	assert.Equal(t, "name.go", filename("name", Config{fIP: false, fTO: false}))
}

func TestFilenameMockOnly(t *testing.T) {
	assert.Equal(t, "mock_name.go", filename("name", Config{fIP: true, fTO: false}))
}

func TestFilenameMockTest(t *testing.T) {
	assert.Equal(t, "mock_name_test.go", filename("name", Config{fIP: true, fTO: true}))
}

func TestFilenameTest(t *testing.T) {
	assert.Equal(t, "name.go", filename("name", Config{fIP: false, fTO: true}))
}
