package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockConstructorReportsCaller(t *testing.T) {
	const childProcess = "MOCKERY_TEST_CONSTRUCTOR_HELPER"
	if os.Getenv(childProcess) == "1" {
		_, file, line, ok := runtime.Caller(0)
		require.True(t, ok)
		t.Logf("expected location: %s:%d", filepath.Base(file), line+3)
		m := NewMockRequester(t)
		m.EXPECT().Get("missing").Return("", nil).Once()
		return
	}

	executable, err := os.Executable()
	require.NoError(t, err)
	cmd := exec.Command(executable, "-test.run=^TestMockConstructorReportsCaller$", "-test.v")
	cmd.Env = append(os.Environ(), childProcess+"=1")
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr, "%s", out)
	require.Equal(t, 1, exitErr.ExitCode(), "%s", out)

	location := regexp.MustCompile(`expected location: (constructor_helpers_test\.go:\d+)`).FindSubmatch(out)
	require.Len(t, location, 2, "%s", out)
	assert.Contains(t, string(out), string(location[1])+": FAIL: 0 out of 1 expectation(s) were met.")
}

func TestMockConstructorWithoutHelper(t *testing.T) {
	for _, fulfilled := range []bool{false, true} {
		t.Run(fmt.Sprintf("fulfilled=%t", fulfilled), func(t *testing.T) {
			legacy := &constructorTestingT{}
			m := NewMockRequester(legacy)
			m.EXPECT().Get("request").Return("response", nil).Once()
			if fulfilled {
				response, err := m.Get("request")
				require.NoError(t, err)
				assert.Equal(t, "response", response)
			}

			require.NotNil(t, legacy.cleanup)
			legacy.cleanup()
			if fulfilled {
				assert.Empty(t, legacy.errors)
			} else {
				require.Len(t, legacy.errors, 1)
				assert.Contains(t, legacy.errors[0], "0 out of 1 expectation(s) were met")
			}
		})
	}
}

func TestMockConstructorTypeParams(t *testing.T) {
	m := NewMockConstructorTypeParams[string, bool](t)
	m.EXPECT().Get().Return("response").Once()
	assert.Equal(t, "response", m.Get())
}

type constructorTestingT struct {
	cleanup func()
	errors  []string
}

func (t *constructorTestingT) Cleanup(f func()) {
	t.cleanup = f
}

func (t *constructorTestingT) Logf(string, ...any) {}

func (t *constructorTestingT) Errorf(format string, args ...any) {
	t.errors = append(t.errors, fmt.Sprintf(format, args...))
}

func (t *constructorTestingT) FailNow() {
	panic("unexpected FailNow")
}
