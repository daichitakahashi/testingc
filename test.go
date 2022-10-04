package testingc

import (
	"bytes"
	"fmt"
	"testing"
)

type T struct {
	C
	out *bytes.Buffer
}

func (t *T) Error(args ...any) {
	t.Log(args...)
	t.Fail()
}

func (t *T) Errorf(format string, args ...any) {
	t.Logf(format, args...)
	t.Fail()
}

func (t *T) Fatal(args ...any) {
	t.Log(args...)
	t.FailNow()
}

func (t *T) Fatalf(format string, args ...any) {
	t.Logf(format, args...)
	t.FailNow()
}

func (t *T) Log(args ...any) {
	_, _ = fmt.Fprintln(t.out, args...)
}

func (t *T) Logf(format string, args ...any) {
	_, _ = fmt.Fprintf(t.out, format, args...)
	t.out.WriteByte('\n')
}

func (t *T) Name() string {
	return "testingc.Test"
}

func (t *T) Skip(args ...any) {
	t.Log(args...)
	t.SkipNow()
}

func (t *T) Skipf(format string, args ...any) {
	t.Logf(format, args...)
	t.SkipNow()
}

var _ testing.TB = (*T)(nil)

type TestResult struct {
	t *T
}

func (r *TestResult) Failed() bool {
	return r.t.Failed()
}

func (r *TestResult) Skipped() bool {
	return r.t.Skipped()
}

func (r *TestResult) Error() string {
	return "" // TODO: record annotated log
}

func (r *TestResult) Logs() []byte {
	return r.t.out.Bytes()
}

func Test(fn func(t *T)) *TestResult {
	t := &T{
		out: bytes.NewBuffer(nil),
	}

	done := make(chan struct{})
	go func() {
		defer func() {
			defer close(done)
			r := recover()
			if r != nil {
				t.Error(r)
			}
			recovered := t.teardown(true)
			if recovered != nil {
				t.Error(recovered)
			}
		}()
		fn(t)
	}()
	<-done

	return &TestResult{
		t: t,
	}
}
