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
// In functions where more than a few errors need to be handled and where all
// errors will be handled in the same way, the handle package can be used to
// ensure consistency while reducing the amount of code dedicated to error
// handling.
//
// If the enclosing function expects to return an error it must use named
// return values so that the check and handle functions can be bound to the
// error value.
//
// The error returned can be wrapped:
//
//     func do(name string) (err error) {
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
//     func do(name string) (err error) {
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
// An enclosing function can use check to trigger an early return with shared
// behavior on errors:
//
//     func do(name string) (err error) {
//         check, handle := handle.Error(&err, func(){
//             // Log err.
//         })
//         defer handle()
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
//         check, handle := handle.Error(&err, func(){
//             // Log err.
//         })
//         defer handle()
//
//         //...
//
//         return
//     }
package handle

import "fmt"

// Error returns a check and handle function. When passed a non-nil error,
// check triggers the deferred handle function to call each function in fns
// before returning from the enclosing function. If err is nil, the check
// and handle functions will be bound to an internal shared error value.
func Error(err *error, fns ...func()) (func(error), func()) {
	var shared error
	if err == nil {
		err = &shared
	}

	return check(err), handle(err, fns...)
}

// Errorf returns a check and handle function. When passed a non-nil error,
// check triggers the deferred handle function to wrap the error being returned
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
