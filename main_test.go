package testingc

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daichitakahashi/gocmd"
)

func execTest(t *testing.T, name string) (string, int) {
	t.Helper()
	goCmd, _, err := gocmd.DetermineFromModuleGoVersion(gocmd.ModeLatest)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(goCmd, "test", "-v", "-count=1",
		"./"+filepath.Join("testdata", "main", name))
	cmd.Env = append(os.Environ(), "TEST_NAME="+t.Name())
	output, err := cmd.Output()

	var code int
	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			t.Fatal(err)
		}
		code = ee.ExitCode()
	}
	var out string
	if len(output) > 0 {
		out = string(output)
		t.Log(out)
	}
	return out, code
}

func requireSuccess(t *testing.T, code int) {
	t.Helper()
	if code != 0 {
		t.Fatalf("unexpected fail: %d", code)
	}
}

func requireFail(t *testing.T, code int) {
	t.Helper()
	if code == 0 {
		t.Fatalf("unexpected sucess: %d", code)
	}
}

func logContains(t *testing.T, logs, message string) {
	t.Helper()
	if !strings.Contains(logs, message) {
		t.Fatalf("log doesn't contain %q", message)
	}
}

func logNotContains(t *testing.T, logs, message string) {
	t.Helper()
	if strings.Contains(logs, message) {
		t.Fatalf("log contains %q unexpectedly", message)
	}
}

func TestMainFunc(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		test func(t *testing.T, logs string, code int)
	}{
		{
			name: "skip",
			test: func(t *testing.T, logs string, code int) {
				requireSuccess(t, code)
				logContains(t, logs, t.Name())
				logNotContains(t, logs, "this test always fails")
			},
		}, {
			name: "skipOnCleanup",
			test: func(t *testing.T, logs string, code int) {
				requireSuccess(t, code)
				logContains(t, logs, t.Name())
				logContains(t, logs, "test passed")
			},
		}, {
			name: "fail",
			test: func(t *testing.T, logs string, code int) {
				requireFail(t, code)
				logContains(t, logs, "FAIL")
				logNotContains(t, logs, "test passed")
			},
		}, {
			name: "failOnCleanup",
			test: func(t *testing.T, logs string, code int) {
				requireFail(t, code)
				logContains(t, logs, "FAIL")
				logContains(t, logs, "test passed")
			},
		}, {
			name: "failNow",
			test: func(t *testing.T, logs string, code int) {
				requireFail(t, code)
				logContains(t, logs, "FAIL")
				logNotContains(t, logs, "test passed")
			},
		}, {
			name: "failNowOnCleanup",
			test: func(t *testing.T, logs string, code int) {
				requireFail(t, code)
				logContains(t, logs, "FAIL")
				logContains(t, logs, "test passed")
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			logs, code := execTest(t, tc.name)
			tc.test(t, logs, code)
		})
	}
}
