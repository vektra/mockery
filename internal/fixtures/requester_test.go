package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestRequesterMock(t *testing.T) {
	m := NewMockRequester(t)
	m.EXPECT().Get("foo").Return("bar", nil).Once()
	retString, err := m.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", retString)
}

func TestRequesterMockRunAndReturn(t *testing.T) {
	m := NewMockRequester(t)
	m.EXPECT().Get(mock.Anything).RunAndReturn(func(path string) (string, error) {
		return path + " world", nil
	})
	retString, err := m.Get("hello")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", retString)
}

func TestRequesterMockRun(t *testing.T) {
	m := NewMockRequester(t)
	m.EXPECT().Get(mock.Anything).Return("", nil)
	m.EXPECT().Get(mock.Anything).Run(func(path string) {
		fmt.Printf("Side effect! Argument is: %s", path)
	})
	retString, err := m.Get("hello")
	assert.NoError(t, err)
	assert.Equal(t, "", retString)
}

//nolint:errcheck
func TestRequesterMockTestifyEmbed(t *testing.T) {
	m := NewMockRequester(t)
	m.EXPECT().Get(mock.Anything).Return("", nil).Twice()
	m.Get("hello")
	m.Get("world")
	assert.Len(t, m.Mock.Calls, 2)
}

func TestRequesterMoq(t *testing.T) {
	m := &MoqRequester{
		GetFunc: func(path string) (string, error) {
			fmt.Printf("Go path: %s\n", path)
			return path + "/foo", nil
		},
	}
	result, err := m.Get("/path")
	assert.NoError(t, err)
	assert.Equal(t, "/path/foo", result)
}
