package testingc

import (
	"testing"
)

type MC struct {
	m *testing.M
	C
}

func (c *MC) Run() (code int) {
	return c.m.Run()
}

func (c *MC) Name() string {
	return "testingc.MC"
}

func M(m *testing.M, fn func(c *MC) int) int {
	c := &MC{
		m: m,
		C: C{
			panicOnFail: true,
		},
	}
	done := make(chan int)

	go func() {
		var s int
		defer func() {
			defer func() {
				// handle failed/skipped
				// for the case that tests have failed/skipped in cleanup functions,
				// we use "defer" in "defer"
				c.smu.Lock()
				status := c.status
				c.smu.Unlock()
				switch {
				case status&failed == failed: // failure is
					s = 1
				case status&skipped == skipped:
					s = 0
				default:
				}
				done <- s
			}()

			// cleanup
			c.teardown()
		}()
		s = fn(c)
	}()
	return <-done
}
