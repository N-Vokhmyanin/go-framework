package errors

import (
	"fmt"
	"github.com/cockroachdb/errors/errbase"
	"runtime"
)

// Stack represents a stack of program counters.
// This mirrors the (non-exported) type of the same name in github.com/pkg/errors.
type Stack []uintptr

// Format mirrors the code in github.com/pkg/errors.
func (s *Stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := errbase.StackFrame(pc)
				_, _ = fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

// StackTrace mirrors the code in github.com/pkg/errors.
func (s *Stack) StackTrace() errbase.StackTrace {
	f := make([]errbase.StackFrame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = errbase.StackFrame((*s)[i])
	}
	return f
}

// Callers mirrors the code in github.com/pkg/errors,
// but makes the depth customizable.
func Callers(depth int) *Stack {
	const numFrames = 32
	var pcs [numFrames]uintptr
	n := runtime.Callers(2+depth, pcs[:])
	var st Stack = pcs[0:n]
	return &st
}
