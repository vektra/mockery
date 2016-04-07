package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"strings"
)

func configFromCommandLine(str string) Config {
	return parseConfigFromArgs(strings.Split(str, " "))
}

func TestParseConfigDefaults(t *testing.T) {
	config := configFromCommandLine("mockery")
	assert.Equal(t, "", config.fName)
	assert.Equal(t, false, config.fPrint)
	assert.Equal(t, "./mocks", config.fOutput)
	assert.Equal(t, ".", config.fDir)
	assert.Equal(t, false, config.fRecursive)
	assert.Equal(t, false, config.fAll)
	assert.Equal(t, false, config.fIP)
	assert.Equal(t, false, config.fTO)
	assert.Equal(t, "camel", config.fCase)
	assert.Equal(t, "", config.fNote)
}

func TestParseConfigFlippingValues(t *testing.T) {
	config := configFromCommandLine("mockery -name hi -print -output output -dir dir -recursive -all -inpkg -testonly -case case -note note")
	assert.Equal(t, "hi", config.fName)
	assert.Equal(t, true, config.fPrint)
	assert.Equal(t, "output", config.fOutput)
	assert.Equal(t, "dir", config.fDir)
	assert.Equal(t, true, config.fRecursive)
	assert.Equal(t, true, config.fAll)
	assert.Equal(t, true, config.fIP)
	assert.Equal(t, true, config.fTO)
	assert.Equal(t, "case", config.fCase)
	assert.Equal(t, "note", config.fNote)
}
