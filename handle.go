// Package handle reduces the boilerplate required for some error handling
// patterns by (ab)using panic and recover for flow control. See the warnings
// section for more detail.
//
// In functions where the handling of each error is unique or where only a
// few errors need to be handled, it is best to handle errors with a simple
// if statement:
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
// a named return value so that the escape.On method and hatch function can
// be bound to it.
//
// The error returned can be wrapped:
//
//     func do(name string) (err error) {
//         escape, hatch := handle.Errorf(&err, "do(%s)", name)
//         defer hatch()
//
//         // ...
//
//         return
//     }
//
// or returned unmodified:
//
//     func do(name string) (err error) {
//         escape, hatch := handle.Error(&err)
//         defer hatch()
//
//         // ...
//
//         return
//     }
//
// With a deferred hatch any call to escape.On with a non-nil error will cause
// the enclosing function to return.
//
//     // Return if err is not nil.
//     f, err := os.Open(name)
//     escape.On(err)
//
// An enclosing function can use escape.On to trigger an early return with
// shared behavior on errors:
//
//     func do(name string) (err error) {
//         escape, hatch := handle.Error(&err, func(){
//             // Log err.
//         })
//         defer hatch()
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
//         escape, hatch := handle.Error(&err, func(){
//             // Log err.
//         })
//         defer hatch()
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
//         escape, hatch := handle.Errorf(&err, "copy %s %s", src, dst)
//         defer hatch()
//
//         r, err := os.Open(src)
//         escape.On(err)
//
//         defer r.Close()
//
//         w, err := os.Create(dst)
//         escape.On(err)
//
//         defer handle.Chain(&err, func() {
//             w.Close()
//             os.Remove(dst)
//         })
//
//         _, err = io.Copy(w, r)
//         escape.On(err)
//
//         return w.Close()
//     }
//
// WARNINGS
//
// Care must be taken to ensure that the hatch function returned by Error
// or Errorf is deferred. Failure to do so will result in an unhandled panic
// when escape.On is invoked.
//
// The escape.On method must not cross goroutine boundaries. In addition, it
// should only be invoked by the the function that deferred the hatch function
// or a function called by that function. Calling the escape.On method in
// another goroutine or after the function that deferred hatch has returned
// will result in an unhandled panic.
//
// If you are unsure, set Name to the name given to the escape object and run:
//
//     Name=escape
//     go build -gcflags '-m' 2>&1 | grep -F "${Name}.On escapes to heap"
//
// If you see:
//
//     path.go:line:column: ${Name}.On escapes to heap
//
// You are probably doing something that won't end well.
package handle

import "fmt"

// Chain adds an additional action, fn, to perform when a non-nil error is
// being returned. Chain must be deferred.
func Chain(err *error, fn func()) {
	if *err != nil {
		fn()
	}
}

// Error returns an escape object and a hatch function. When passed a non-nil
// error, escape.On sets the bound error *err and triggers the deferred hatch
// function to recover the panic (if there was one) and then, while *err
// remains non-nil, it calls each function in fns (in reverse order to match
// the LIFO order of deferred functions).
func Error(err *error, fns ...func()) (*escape, func()) {
	var shared error

	if err == nil {
		err = &shared
	}

	s := &escape{err: err, fns: fns}

	return s, func() {
		if s.pnc {
			s.pnc = false

			_ = recover()
		}

		// Call the error functions in while *s.err is not nil.
		// Functions are called in reverse order to match defers.
		for i := len(s.fns) - 1; *s.err != nil && i >= 0; i-- {
			fns[i]()
		}
	}
}

// Errorf calls Error passing it a function that wraps the error returned.
func Errorf(err *error, format string, args ...interface{}) (*escape, func()) {
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

type escape struct {
	err *error
	fns []func()
	pnc bool
}

// On sets the bound error to the error passed if that error is non-nil and
// then triggers a panic if one hasn't already been triggered.
func (s *escape) On(ce error) {
	if ce != nil {
		*s.err = ce

		// Only panic if we haven't previously.
		if !s.pnc {
			s.pnc = true

			panic(failure{ce})
		}
	}
}
