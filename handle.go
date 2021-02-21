// Package handle reduces the boilerplate required for some error handling
// patterns.
//
// In functions where the handling of each error is unique or where only a
// few errors need to be handled, it is probably best to handle errors the
// standard way:
//
//     f, err := os.Open(name)
//     if err != nil {
//         // Handle error.
//     }
//
// And, of course, the usual advice about treating errors as values and using
// the full power of Go to simplify error handling applies.
//
// In functions where more than a few errors need to be handled and where all
// errors will be handled in a similar way, the handle package can be used to
// ensure consistency while reducing the amount of code dedicated to error
// handling.
//
// If the enclosing function expects to return an error, that error must be
// a named return value so that the check and done functions can be bound to
// it.
//
// The error returned can be wrapped:
//
//     func do(name string) (err error) {
//         check, done := handle.Errorf(&err, "do(%s)", name)
//         defer done()
//
//         // ...
//
//         return
//     }
//
// or returned unmodified:
//
//     func do(name string) (err error) {
//         check, done := handle.Error(&err)
//         defer done()
//
//         // ...
//
//         return
//     }
//
// With a deferred done any call to check with a non-nil error will cause the
// enclosing function to return.
//
//     // Return if err is not nil.
//     f, err := os.Open(name); check(err)
//
// An enclosing function can use check to trigger an early return with shared
// behavior on errors:
//
//     func do(name string) (err error) {
//         check, done := handle.Error(&err, func(){
//             // Log err.
//         })
//         defer done()
//
//         //...
//
//         return
//     }
//
// and it can do so even if the enclosing function does not return an error:
//
//     func do(name string) {
//         var err error
//         check, done := handle.Error(&err, func(){
//             // Log err.
//         })
//         defer done()
//
//         //...
//
//         return
//     }
//
// Additional error handling actions can be added with handle.Chain as in
// the example below adapted from Error Handling - Problem Overview:
//
// github.com/golang/proposal/blob/master/design/go2draft-error-handling-overview.md
//
//     func CopyFile(src, dst string) (err error) {
//         check, done := handle.Errorf(&err, "copy %s %s", src, dst)
//         defer done()
//
//         r, err := os.Open(src); check(err)
//         defer r.Close()
//
//         w, err := os.Create(dst); check(err)
//         defer handle.Chain(&err, func() {
//             w.Close()
//             os.Remove(dst)
//         })
//
//         _, err = io.Copy(w, r); check(err)
//         return w.Close()
//     }
package handle

import "fmt"

// Chain adds an additional action, fn, to perform when the return of a non-nil
// error is triggered by check or by a regular return. Chain must be deferred.
func Chain(err *error, fn func()) {
	if f, ok := (*err).(failure); ok { //nolint:errorlint
		*err = f.error

		// Rather than restoring *err, if we have to unwrap it, we re-wrap it,
		// in case the function fn changes err.
		defer func() {
			*err = failure{*err}
		}()
	}

	if *err != nil {
		fn()
	}
}

// Error returns a check and done function. When passed a non-nil error,
// check triggers the deferred done function to call each function in fns
// before returning from the enclosing function. If err is nil, the check
// and done functions will be bound to an internal shared error value.
func Error(err *error, fns ...func()) (func(error), func()) {
	var shared error
	if err == nil {
		err = &shared
	}

	panicking := false

	return check(&panicking, err), done(&panicking, err, fns...)
}

// Errorf returns a check and done function. When passed a non-nil error,
// check triggers the deferred done function to wrap the error being returned
// from the enclosing function.
func Errorf(err *error, format string, args ...interface{}) (func(error), func()) {
	return Error(err, func() {
		*err = fmt.Errorf(format+": %w", append(args, *err)...) //nolint:goerr113
	})
}

type failure struct {
	error
}

// Error reports the failure as unhandled when encountered "in the wild".
func (f failure) Error() string {
	s := "unhandled error"
	if f.error != nil {
		s += ": " + f.error.Error()
	}

	return s
}

func check(panicking *bool, err *error) func(error) {
	return func(ce error) {
		if ce != nil {
			*err = failure{ce}

			// Only panic if we haven't previously.
			if !*panicking {
				*panicking = true

				panic(*err)
			}
		}
	}
}

func done(panicking *bool, err *error, fns ...func()) func() {
	return func() {
		if *panicking {
			*panicking = false

			_ = recover()

			if f, ok := (*err).(failure); ok { //nolint:errorlint
				*err = f.error
			}
		}

		// If *err was set by check or normal return, call the error functions.
		if *err != nil {
			for _, fn := range fns {
				fn()
			}
		}
	}
}
