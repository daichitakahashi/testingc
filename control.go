package testingc

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// status
const (
	failed  = 1
	skipped = 2
)

type C struct {
	testing.TB
	status    int32
	m         sync.Mutex
	cleanup   []func()
	tmpDir    string
	tmpDirSeq int32
}

func (c *C) Cleanup(f func()) {
	c.m.Lock()
	defer c.m.Unlock()
	c.cleanup = append(c.cleanup, f)
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
	atomic.CompareAndSwapInt32(&c.status, 0, failed)
}

func (c *C) FailNow() {
	c.Fail()
	panic("failed")
}

func (c *C) Failed() bool {
	return atomic.LoadInt32(&c.status) == failed
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
	err := os.Setenv(key, value)
	if err != nil {
		c.Fatal(err)
	}
	c.Cleanup(func() {
		_ = os.Unsetenv(key)
	})
}

func (c *C) Skip(args ...any) {
	log.Println(args...)
	c.SkipNow()
}

func (c *C) SkipNow() {
	atomic.CompareAndSwapInt32(&c.status, 0, skipped)
	panic("skipped")
}

func (c *C) Skipf(format string, args ...any) {
	log.Printf(format, args...)
	c.SkipNow()
}

func (c *C) Skipped() bool {
	return atomic.LoadInt32(&c.status) == skipped
}

func (c *C) TempDir() string {
	c.m.Lock()
	defer c.m.Unlock()

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