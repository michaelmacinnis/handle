// Package handle reduces the amount of boilerplate required to handle errors.
//
// To use handle, the enclosing function must use named return values. The
// error returned can be wrapped:
//
//     do(name string) (err error) {
//         check, handle := handle.Errorf(&err, "do(%s)", name); defer handle()
//
//         // ...
//
//         return
//     }
//
// or returned unmodified:
//
//     do(name string) (err error) {
//         check, handle := handle.Error(&err); defer handle()
//
//         // ...
//
//         return
//     }
//
// With a deferred handle any call to `check` with a non-nil error will cause
// the enclosing function to return.
//
//     // Return if err is not nil.
//     f, err := os.Open(name); check(err)
//
package handle

import "fmt"

func Error(err *error) (func(error), func()) {
	return errorf(unwrapped, err, "", nil)
}

func Errorf(err *error, format string, args ...interface{}) (func(error), func()) {
	return errorf(rewrapped, err, format, args)
}

type failure struct {
	error
}

type wrapper func(failure, string, ...interface{}) error

func errorf(w wrapper, err *error, format string, args ...interface{}) (func(error), func()) {
	return func(ce error) {
			if ce != nil {
				*err = failure{ce}
				panic(*err)
			}
		}, func() {
			if f, ok := (*err).(failure); ok { //nolint:errorlint
				*err = w(f, format, args)

				r := recover()
				if rf, ok := r.(failure); !ok || rf != f {
					panic("unexpected error")
				}
			}
		}
}

func rewrapped(f failure, format string, args ...interface{}) error {
	return fmt.Errorf(format+": %w", append(args, f.error)...) //nolint:goerr113
}

func unwrapped(f failure, format string, args ...interface{}) error {
	return f.error
}
