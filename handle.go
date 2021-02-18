// Package handle reduces the boilerplate required for some error handling
// patterns.
//
// The enclosing function must use named return values. The error returned
// can be wrapped:
//
//     do(name string) (err error) {
//         check, handle := handle.Errorf(&err, "do(%s)", name)
//         defer handle()
//
//         // ...
//
//         return
//     }
//
// or returned unmodified:
//
//     do(name string) (err error) {
//         check, handle := handle.Error(&err)
//         defer handle()
//
//         // ...
//
//         return
//     }
//
// With a deferred handle any call to check with a non-nil error will cause
// the enclosing function to return.
//
//     // Return if err is not nil.
//     f, err := os.Open(name); check(err)
//
package handle

import "fmt"

// Error returns a check and handle function. When passed a non-nil error,
// check triggers the deferred handle function to call each functions in fns
// before returning from the enclosing function.
func Error(err *error, fns ...func()) (func(error), func()) {
	if err == nil {
		var shared error

		return check(&shared), handle(&shared, fns...)
	}

	return check(err), handle(err, fns...)
}

// Errorf returns a check and handle function. When passed a non-nil error,
// check triggers the deferred handle function to wrap and return the error
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

func check(err *error) func(error) {
	return func(ce error) {
		if ce != nil {
			*err = failure{ce}
			panic(*err)
		}
	}
}

func handle(err *error, fns ...func()) func() {
	return func() {
		if f, ok := (*err).(failure); ok { //nolint:errorlint
			// We should be here because of a call to check so recover the panic.
			r := recover()
			if rf, ok := r.(failure); !ok || rf != f {
				// If we don't recover the same error something went very wrong...
				panic(fmt.Sprintf("unexpected panic %v, while handling %v", r, f.error))
			}

			*err = f.error
		}

		// If *err is set either by check or a normal return, call the error functions.
		if *err != nil {
			for _, fn := range fns {
				fn()
			}
		}
	}
}
