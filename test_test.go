package testingc_test

import (
	"strings"
	"testing"

	"github.com/daichitakahashi/testingc"
)

func checkStatus(t *testing.T, r *testingc.TestResult, failed, skipped bool, cause string) {
	t.Helper()

	if failed != r.Failed() {
		if failed {
			t.Error("test is expected to success, but failed")
		} else {
			t.Error("test is expected to fail, but succeeded")
		}
	}
	if skipped != r.Skipped() {
		if skipped {
			t.Error("test is expected to be skipped, but not")
		} else {
			t.Error("test is expected not to be skipped, but skipped")
		}
	}
	if r.Error() != cause { // TODO: compare with annotated error
		t.Logf("want cause: %s", cause)
		t.Errorf("got cause: %s", r.Error())
	}
	if t.Failed() {
		t.FailNow()
	}
}

func checkLog(t *testing.T, want string, got []byte, contains bool) {
	t.Helper()

	if contains {
		if !strings.Contains(string(got), want) {
			t.Logf("wanted log message not found:")
			t.Logf("want log: %s", want)
			t.Fatalf("got log: %s", got)
		}
	} else if want != string(got) {
		t.Logf("want log: %s", want)
		t.Fatalf("got log: %s", got)
	}
}

func TestTest_Fail(t *testing.T) {
	t.Parallel()

	r := testingc.Test(func(c *testingc.T) {
		c.Fail()
	})
	checkStatus(t, r, true, false, "")
	checkLog(t, "", r.Logs(), false)
}

func TestTest_FailNow(t *testing.T) {
	t.Parallel()

	r := testingc.Test(func(c *testingc.T) {
		c.FailNow()
		c.Skip("skip")
	})
	checkStatus(t, r, true, false, "")
	checkLog(t, "", r.Logs(), false)
}

func TestTest_SkipNow(t *testing.T) {
	t.Parallel()

	r := testingc.Test(func(c *testingc.T) {
		c.SkipNow()
	})
	checkStatus(t, r, false, true, "")
	checkLog(t, "", r.Logs(), false)
}

func TestTest_SkipNowAfterFail(t *testing.T) {
	t.Parallel()

	r := testingc.Test(func(c *testingc.T) {
		c.Fail()
		c.SkipNow()
	})
	checkStatus(t, r, true, true, "")
	checkLog(t, "", r.Logs(), false)
}
