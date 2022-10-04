package testingc

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// status
const (
	failed = 1 << iota
	skipped
)

type C struct {
	testing.TB

	smu    sync.Mutex
	status int8

	cmu     sync.Mutex
	cleanup []func()

	tmpDir      string
	tmpDirSeq   int32
	panicOnFail bool
}

func (c *C) Cleanup(f func()) {
	c.cmu.Lock()
	defer c.cmu.Unlock()
	c.cleanup = append(c.cleanup, f)
}

func (c *C) teardown(recoverPanic bool) (recovered any) {
	defer func() {
		c.cmu.Lock()
		remain := len(c.cleanup) > 0
		c.cmu.Unlock()
		if remain {
			r := c.teardown(recoverPanic)
			if recovered == nil {
				recovered = r
			}
		}
	}()
	if recoverPanic {
		defer func() {
			recovered = recover()
		}()
	}

	// extract and run
	for {
		var cleanup func()
		c.cmu.Lock()
		if l := len(c.cleanup); l > 0 {
			cleanup = c.cleanup[l-1]
			c.cleanup = c.cleanup[:l-1]
		}
		c.cmu.Unlock()
		if cleanup == nil {
			break
		}
		cleanup()
	}
	return nil
}

func (c *C) Error(args ...any) {
	log.Println(args...)
	c.Fail()
}

func (c *C) Errorf(format string, args ...any) {
	log.Printf(format, args...)
	c.Fail()
}

func (c *C) Fail() {
	c.smu.Lock()
	defer c.smu.Unlock()
	c.status |= failed
}

func (c *C) FailNow() {
	c.smu.Lock()
	defer c.smu.Unlock()
	c.status |= failed

	if c.panicOnFail {
		panic("FAIL")
	}
	runtime.Goexit()
}

func (c *C) Failed() bool {
	c.smu.Lock()
	defer c.smu.Unlock()
	return c.status&failed == failed
}

func (c *C) Fatal(args ...any) {
	log.Println(args...)
	c.FailNow()
}

func (c *C) Fatalf(format string, args ...any) {
	log.Printf(format, args...)
	c.FailNow()
}

func (*C) Helper() {}

func (*C) Log(args ...any) {
	log.Println(args...)
}

func (*C) Logf(format string, args ...any) {
	log.Printf(format, args...)
}

func (*C) Name() string {
	return "testingc.C"
}

func (c *C) Setenv(key, value string) {
	prev, ok := os.LookupEnv(key)

	err := os.Setenv(key, value)
	if err != nil {
		c.Fatal(err)
	}

	if ok {
		c.Cleanup(func() {
			_ = os.Setenv(key, prev)
		})
	} else {
		c.Cleanup(func() {
			_ = os.Unsetenv(key)
		})
	}
}

func (c *C) Skip(args ...any) {
	log.Println(args...)
	c.SkipNow()
}

func (c *C) SkipNow() {
	c.smu.Lock()
	defer c.smu.Unlock()
	c.status |= skipped
	runtime.Goexit()
}

func (c *C) Skipf(format string, args ...any) {
	log.Printf(format, args...)
	c.SkipNow()
}

func (c *C) Skipped() bool {
	c.smu.Lock()
	defer c.smu.Unlock()
	return c.status&skipped == skipped
}

func (c *C) TempDir() string {
	c.cmu.Lock()
	defer c.cmu.Unlock()

	var exists bool
	if c.tmpDir != "" {
		_, err := os.Stat(c.tmpDir)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			c.Fatalf("TempDir: %v", err)
		}
		exists = err == nil
	}

	if !exists {
		var err error
		c.tmpDir, err = os.MkdirTemp("", "")
		if err != nil {
			c.Fatalf("TempDir: %v", err)
		}
		c.cleanup = append(c.cleanup, func() {
			err := removeAll(c.tmpDir)
			if err != nil {
				c.Errorf("TempDir cleanup: %v", err)
			}
		})
	}

	name := fmt.Sprintf("%s%c%03d", c.tmpDir, filepath.Separator, c.tmpDirSeq)
	c.tmpDirSeq++
	err := os.Mkdir(name, 0777)
	if err != nil {
		c.Fatalf("TempDir: %v", err)
	}
	return name
}

var _ testing.TB = (*C)(nil)

func removeAll(path string) error {
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	var err error
	for {
		select {
		case <-timer.C:
			return err
		default:
			err = os.RemoveAll(path)
			if err == nil {
				return nil
			}
		}
	}
}

var mu sync.Mutex

func output(s string, skip int) {
	mu.Lock()
	defer mu.Unlock()

	buf := new(strings.Builder)
	buf.WriteString("    ")
	_, filename, line, ok := runtime.Caller(skip)
	if ok {
		_, _ = fmt.Fprintf(buf, "%s:%d: ", filepath.Base(filename), line)
	} else {
		buf.WriteString("???: ")
	}

	l := len(s)
	if s[l-1] == '\n' {
		s = s[:l-1]
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if i > 0 {
			buf.WriteString("\n    ")
		}
		buf.WriteString(line)
	}
	buf.WriteByte('\n')
	fmt.Print(buf.String())
}
