// Package exception provides a simple exception handling mechanism for Go,
// simulating try/catch/finally behavior using panic and recover.
// It captures detailed error information including a formatted stack trace.
package exception

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// Exception represents an error with a message, an optional inner error,
// a stack trace capturing the chain of function calls, and a timestamp.
type Exception struct {
	// Message contains the error message.
	Message string
	// InnerError holds an underlying error that caused the exception, if any.
	InnerError error
	// StackTrace contains the formatted stack trace from the point where the exception was created.
	StackTrace string
	// Timestamp is the time when the exception was created.
	Timestamp time.Time
}

// Error implements the built-in error interface.
// If an inner error exists, it appends the inner error's message.
func (e *Exception) Error() string {
	if e.InnerError != nil {
		return fmt.Sprintf("[%s] %s -- caused by: %s",
			e.Timestamp.Format(time.RFC3339), e.Message, e.InnerError.Error())
	}
	return fmt.Sprintf("[%s] %s", e.Timestamp.Format(time.RFC3339), e.Message)
}

// FullDetails returns the complete details of the exception including
// the timestamp, message, inner error (if any), and the full stack trace.
func (e *Exception) FullDetails() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "Timestamp: %s\n", e.Timestamp.Format(time.RFC3339))
	_, _ = fmt.Fprintf(&b, "Message: %s\n", e.Message)
	if e.InnerError != nil {
		_, _ = fmt.Fprintf(&b, "Inner Error: %v\n", e.InnerError)
	}
	_, _ = fmt.Fprintf(&b, "Stack Trace:\n%s\n", e.StackTrace)
	return b.String()
}

// DumpToFile writes the full exception details to the specified file.
// It returns an error if the write operation fails.
func (e *Exception) DumpToFile(filename string) error {
	return os.WriteFile(filename, []byte(e.FullDetails()), 0644)
}

// New creates a new Exception with the given message. The function captures
// the current time and a stack trace starting from a specified number of frames to skip.
func New(msg string) *Exception {
	return &Exception{
		Message:    msg,
		Timestamp:  time.Now(),
		StackTrace: captureStackTrace(3),
	}
}

// NewFromError creates a new Exception from an existing error by wrapping it.
// The resulting Exception's message is derived from the original error.
func NewFromError(err error) *Exception {
	return New(err.Error()).WithInnerError(err)
}

// WithInnerError attaches an inner error to the Exception, allowing error chaining.
// It returns the modified Exception to allow method chaining.
func (e *Exception) WithInnerError(err error) *Exception {
	e.InnerError = err
	return e
}

// captureStackTrace retrieves a formatted stack trace starting from the given number of frames to skip.
// It filters out any internal frames belonging to the exception package.
func captureStackTrace(skip int) string {
	var sb strings.Builder
	for i := skip; ; i++ {
		// Retrieve the program counter, file, and line number for the caller.
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		// Filter out internal frames from this package based on the function name.
		if isInternal(fn.Name()) {
			continue
		}
		_, _ = fmt.Fprintf(&sb, "    at %s (%s:%d)\n", fn.Name(), file, line)
	}
	return sb.String()
}

// isInternal determines whether the function name belongs to the exception package.
// It splits the function name by "/" and checks if the last part starts with "exception.".
func isInternal(funcName string) bool {
	parts := strings.Split(funcName, "/")
	if len(parts) == 0 {
		return false
	}
	last := parts[len(parts)-1]
	return strings.HasPrefix(last, "exception.")
}

// TryCatch simulates a try/catch block. It executes the try function; if a panic occurs,
// it recovers from the panic and passes an Exception to the catch function.
// It returns the value from try (if successful) or catch (if an exception occurred),
// along with a boolean indicating whether the operation completed without exceptions.
func TryCatch(try func() any, catch func(e *Exception) any) (ret any, ok bool) {
	ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
			var ex *Exception
			switch v := r.(type) {
			case *Exception:
				ex = v
			case error:
				ex = NewFromError(v)
			default:
				ex = New(fmt.Sprintf("%v", v))
			}
			ret = catch(ex)
		}
	}()
	ret = try()
	return ret, ok
}

// TryCatchFinally simulates a try/catch/finally block. It executes the try function,
// and in case of a panic, it recovers and calls the catch function. Regardless of a panic,
// the finally function is always executed.
// It returns the value from try (if successful) or catch (if an exception occurred),
// along with a boolean indicating whether the operation completed without exceptions.
func TryCatchFinally(try func() any, catch func(e *Exception) any, finally func()) (ret any, ok bool) {
	defer finally()
	return TryCatch(try, catch)
}

// TryCatchT is a generic version of TryCatch that allows specifying the return type explicitly.
// It executes the try function; if a panic occurs, it recovers and passes an Exception to the catch function.
// It returns the value from try (if successful) or catch (if an exception occurred) with the specified type T,
// along with a boolean indicating whether the operation completed without exceptions.
func TryCatchT[T any](try func() T, catch func(e *Exception) T) (ret T, ok bool) {
	ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
			var ex *Exception
			switch v := r.(type) {
			case *Exception:
				ex = v
			case error:
				ex = NewFromError(v)
			default:
				ex = New(fmt.Sprintf("%v", v))
			}
			ret = catch(ex)
		}
	}()
	ret = try()
	return ret, ok
}

// TryCatchFinallyT is a generic version of TryCatchFinally that allows specifying the return type explicitly.
// It executes the try function, and in case of a panic, it recovers and calls the catch function.
// Regardless of a panic, the finally function is always executed.
// It returns the value from try (if successful) or catch (if an exception occurred) with the specified type T,
// along with a boolean indicating whether the operation completed without exceptions.
func TryCatchFinallyT[T any](try func() T, catch func(e *Exception) T, finally func()) (ret T, ok bool) {
	defer finally()
	return TryCatchT(try, catch)
}

// Throw panics with a new Exception created with the specified message.
func Throw(msg string) {
	panic(New(msg))
}

// Throwf panics with a new Exception created by formatting the specified message.
func Throwf(format string, args ...interface{}) {
	panic(New(fmt.Sprintf(format, args...)))
}

// ThrowIf panics with an Exception if the specified condition is true.
func ThrowIf(condition bool, msg string) {
	if condition {
		Throw(msg)
	}
}

// The following anonymous variable references all exported symbols to ensure they are used,
// avoiding potential "unused" warnings from certain static analysis tools.
var _ = struct {
	New              func(string) *Exception
	NewFromError     func(error) *Exception
	TryCatch         func(try func() any, catch func(e *Exception) any) (any, bool)
	TryCatchFinally  func(try func() any, catch func(e *Exception) any, finally func()) (any, bool)
	TryCatchT        func(try func() any, catch func(e *Exception) any) (any, bool)
	TryCatchFinallyT func(try func() any, catch func(e *Exception) any, finally func()) (any, bool)
	Throw            func(string)
	Throwf           func(string, ...interface{})
	ThrowIf          func(bool, string)
}{
	New:              New,
	NewFromError:     NewFromError,
	TryCatch:         TryCatch,
	TryCatchFinally:  TryCatchFinally,
	TryCatchT:        TryCatchT[any],
	TryCatchFinallyT: TryCatchFinallyT[any],
	Throw:            Throw,
	Throwf:           Throwf,
	ThrowIf:          ThrowIf,
}
