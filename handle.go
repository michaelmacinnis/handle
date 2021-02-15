// Package handle reduces the boilerplate required for some error handling
// patterns.
//
// To use handle.Errorf or handle.Error, the enclosing function must use named
// return values. The error returned can be wrapped:
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
// check triggers the deferred handle function to return the error from the
// enclosing function.
func Error(err *error) (func(error), func()) {
	return check(err), handle(err, func(ce error) {
		*err = ce
	})
}

// Errorf returns a check and handle function. When passed a non-nil error,
// check triggers the deferred handle function to wrap and return the error
// from the enclosing function.
func Errorf(err *error, format string, args ...interface{}) (func(error), func()) {
	return check(err), handle(err, func(ce error) {
		*err = fmt.Errorf(format+": %w", append(args, ce)...) //nolint:goerr113
	})
}

// Func returns a check and handle function. When passed a non-nil error,
// check triggers the deferred handle function to call the function ef with
// the error before returning from the enclosing function.
func Func(ef func(error)) (func(error), func()) {
	var err error

	return check(&err), handle(&err, ef)
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

func handle(err *error, cb func(error)) func() {
	return func() {
		if f, ok := (*err).(failure); ok { //nolint:errorlint
			r := recover()
			if rf, ok := r.(failure); !ok || rf != f {
				panic(fmt.Sprintf("unexpected panic %v, while handling %v", r, f.error))
			}

			cb(f.error)
		}
	}
}
