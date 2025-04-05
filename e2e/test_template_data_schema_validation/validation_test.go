package test_template_exercise

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/chigopher/pathlib"
	"github.com/stretchr/testify/assert"
)

func TestExercise(t *testing.T) {
	t.Parallel()
	outfile := pathlib.NewPath("./exercise.txt")
	//nolint:errcheck
	defer outfile.Remove()

	out, err := exec.Command(
		"go", "run", "github.com/vektra/mockery/v3",
		"--config", "./.mockery.yml").CombinedOutput()
	assert.Error(t, err)
	expectedString := "ERR (root): foo is required"
	assert.True(t, strings.Contains(string(out), expectedString), "expected string in stdout not found: \"%s\"", expectedString)
}
