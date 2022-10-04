package testingc

import (
	"testing"
)

type M struct {
	m *testing.M
	C
}

func (m *M) checkFailed() {
	m.smu.Lock()
	status := m.status
	m.smu.Unlock()

	if status&failed == failed {
		m.FailNow()
	}
}

func (m *M) Run() (code int) {
	m.checkFailed()
	return m.m.Run()
}

func (m *M) Name() string {
	return "testingc.Main"
}

func Main(m *testing.M, fn func(m *M)) {
	c := &M{
		m: m,
		C: C{
			panicOnFail: true,
		},
	}

	done := make(chan struct{})
	go func() {
		defer func() {
			defer func() {
				// If Main returns before the runtime detects a panic which is
				// not recovered, Main will not fail.
				// To wait for the goroutine that may cause a panic to exit first,
				// schedule the close of `done` in another new gorountine.
				go close(done)
			}()
			c.teardown(false)
			c.checkFailed()
		}()
		fn(c)
	}()
	<-done
}
