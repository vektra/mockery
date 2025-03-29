package test_template_exercise

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/chigopher/pathlib"
	"github.com/stretchr/testify/assert"
)

func TestExercise(t *testing.T) {
	outfile := pathlib.NewPath("./exercise.txt")
	//nolint:errcheck
	defer outfile.Remove()

	out, err := exec.Command(
		"go", "run", "github.com/vektra/mockery/v3",
		"--config", "./.mockery.yml").CombinedOutput()
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(out))
		os.Exit(1)
	}

	b, err := outfile.ReadFile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	expectedPath := pathlib.NewPath("exercise_expected.txt")
	expected, err := expectedPath.ReadFile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	assert.Equal(t, string(expected), string(b))
}
