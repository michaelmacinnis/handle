package handle_test

import (
	"errors"
	"fmt"

	"github.com/michaelmacinnis/handle"
)

func ExampleError_func() {
	var err error
	escape, hatch := handle.Error(&err, func() {
		fmt.Printf("err: %s\n", err.Error())
	})
	defer hatch()

	s, err := works("World!")
	escape.On(err)

	s, err = fails("World!")
	escape.On(err)

	// We will never reach here
	fmt.Printf("%s\n", s)

	// Output: err: failure
}

func ExampleError_unmodified() {
	f := func() (err error) {
		escape, hatch := handle.Error(&err)
		defer hatch()

		s, err := fails("World!")
		escape.On(err)

		// We will never reach here

		s, err = works("World!")
		escape.On(err)

		fmt.Printf("%s\n", s)

		return
	}

	fmt.Printf("%s\n", f().Error())

	// Output: failure
}

func annotate(e *handle.Escape) func(error, string, ...interface{}) {
	return func(err error, format string, args ...interface{}) {
		if err == nil {
			return
		}

		e.On(fmt.Errorf(format+": %w", append(args, err)...))
	}
}

func failure() error {
	return errFailure
}

func additionalContext() (err error) {
	escape, hatch := handle.Errorf(&err, "copy(src, dst)")
	defer hatch()
	check := annotate(escape)

	check(failure(), "call to failure() failed")

	return
}

func Example_annotate() {
	fmt.Printf("%s\n", additionalContext().Error())
	// Output:
	// copy(src, dst): call to failure() failed: failure
}

func ExampleCopyFile_close_dst_err() {
	docopy(map[string]error{
		"close(dst)": errors.New("problem closing dst"),
	})
	// Output:
	// open(src)
	// open(dst)
	// copy(dst, src)
	// close(dst)
	// close(dst)
	// remove(dst)
	// close(src)
	// copy(src, dst): problem closing dst
}

func ExampleCopyFile_copy_err() {
	docopy(map[string]error{
		"copy(dst, src)": errors.New("problem copying"),
	})
	// Output:
	// open(src)
	// open(dst)
	// copy(dst, src)
	// close(dst)
	// remove(dst)
	// close(src)
	// copy(src, dst): problem copying
}

func ExampleCopyFile_no_dst() {
	docopy(map[string]error{
		"open(dst)": errors.New("dst not found"),
	})
	// Output:
	// open(src)
	// open(dst)
	// close(src)
	// copy(src, dst): dst not found
}

func ExampleCopyFile_no_error() {
	docopy(map[string]error{})
	// Output:
	// open(src)
	// open(dst)
	// copy(dst, src)
	// close(dst)
	// close(src)
}

func ExampleCopyFile_no_src() {
	docopy(map[string]error{
		"open(src)": errors.New("src not found"),
	})
	// Output:
	// open(src)
	// copy(src, dst): src not found
}

func ExampleHandleErr() {
	var err error
	escape, hatch := handle.Error(&err, func() {
		fmt.Printf("we should never see this\n")
	})
	defer hatch()

	defer handle.Chain(&err, func() {
		fmt.Printf("we should never see this either\n")
	})

	defer handle.Chain(&err, func() {
		fmt.Printf("error handled\n")

		// Pretend we did something here to handle the error.
		// To stop other handlers for firing we set err to nil.
		err = nil
	})

	escape.On(errors.New("an error"))
	// Output:
	// error handled
}

func docopy(data map[string]error) {
	err := mockcopy(data)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
}

func mock(data map[string]error, name string) error {
	fmt.Printf("%s\n", name)
	return data[name]
}

/*
This function is meant to emulate the CopyFile function from:

https://go.googlesource.com/proposal/+/master/design/go2draft-error-handling-overview.md

	func CopyFile(src, dst string) error {
		handle err {
			return fmt.Errorf("copy %s %s: %v", src, dst, err)
		}

		r := check os.Open(src)
		defer r.Close()

		w := check os.Create(dst)
		handle err {
			w.Close()
			os.Remove(dst) // (only if a check fails)
		}

		check io.Copy(w, r)
		check w.Close()
		return nil
	}
*/
func mockcopy(data map[string]error) (err error) {
	escape, hatch := handle.Errorf(&err, "copy(src, dst)")
	defer hatch()

	err = mock(data, "open(src)")
	escape.On(err)

	defer mock(data, "close(src)")

	err = mock(data, "open(dst)")
	escape.On(err)

	defer handle.Chain(&err, func() {
		mock(data, "close(dst)")
		mock(data, "remove(dst)")
	})

	err = mock(data, "copy(dst, src)")
	escape.On(err)

	return mock(data, "close(dst)")
}
