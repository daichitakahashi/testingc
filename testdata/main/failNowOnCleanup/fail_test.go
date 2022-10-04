package failNowOnCleanup

import (
	"testing"

	"github.com/daichitakahashi/testingc"
)

func TestMain(m *testing.M) {
	testingc.Main(m, func(m *testingc.M) {
		m.Cleanup(func() {
			m.FailNow()
		})
		m.Run()
	})
}

func Test(t *testing.T) {
	t.Log("test passed")
}
