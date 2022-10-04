package skip

import (
	"os"
	"testing"

	"github.com/daichitakahashi/testingc"
)

func TestMain(m *testing.M) {
	testingc.Main(m, func(m *testingc.M) {
		m.Skip(os.Getenv("TEST_NAME"))
		m.Run()
	})
}

func TestFails(t *testing.T) {
	t.Fatal("this test always fails")
}
