package mockery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilenameBare(t *testing.T) {
	out := FileOutputStreamProvider{InPackage: false, TestOnly: false}
	assert.Equal(t, "name.go", out.filename("name"))
}

func TestFilenameMockOnly(t *testing.T) {
	out := FileOutputStreamProvider{InPackage: true, TestOnly: false}
	assert.Equal(t, "mock_name.go", out.filename("name"))
}

func TestFilenameMockTest(t *testing.T) {
	out := FileOutputStreamProvider{InPackage: true, TestOnly: true}
	assert.Equal(t, "mock_name_test.go", out.filename("name"))
}

func TestFilenameTest(t *testing.T) {
	out := FileOutputStreamProvider{InPackage: false, TestOnly: true}
	assert.Equal(t, "name_test.go", out.filename("name"))
}

func TestUnderscoreCaseName(t *testing.T) {
	assert.Equal(t, "notify_event", (&FileOutputStreamProvider{}).underscoreCaseName("NotifyEvent"))
	assert.Equal(t, "repository", (&FileOutputStreamProvider{}).underscoreCaseName("Repository"))
	assert.Equal(t, "http_server", (&FileOutputStreamProvider{}).underscoreCaseName("HTTPServer"))
	assert.Equal(t, "awesome_http_server", (&FileOutputStreamProvider{}).underscoreCaseName("AwesomeHTTPServer"))
	assert.Equal(t, "csv", (&FileOutputStreamProvider{}).underscoreCaseName("CSV"))
	assert.Equal(t, "position0_size", (&FileOutputStreamProvider{}).underscoreCaseName("Position0Size"))
}
