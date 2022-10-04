package fail

import (
	"testing"

	"github.com/daichitakahashi/testingc"
)

func TestMain(m *testing.M) {
	testingc.Main(m, func(m *testingc.M) {
		m.Fail()
		m.Run()
	})
}

func Test(t *testing.T) {
	t.Log("test passed")
}
