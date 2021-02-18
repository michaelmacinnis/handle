package handle_test

import (
	"errors"
	"fmt"

	"github.com/michaelmacinnis/handle"
)

func ExampleError_unmodified() {
	f := func() (err error) {
		check, done := handle.Error(&err)
		defer done()

		s, err := fails("World!")
		check(err)

		// We will never reach here

		s, err = works("World!")
		check(err)

		fmt.Printf("%s\n", s)

		return
	}

	fmt.Printf("%s\n", f().Error())

	// Output: failure
}

func ExampleError_func() {
	var err error
	check, done := handle.Error(&err, func() {
		fmt.Printf("err: %s\n", err.Error())
	})
	defer done()

	s, err := works("World!")
	check(err)

	s, err = fails("World!")
	check(err)

	// We will never reach here
	fmt.Printf("%s\n", s)

	// Output: err: failure
}

func ExampleCopyFile_no_error() {
	dofakecopy(map[string]error{})
	// Output:
	// open(src)
	// open(dst)
	// copy(dst, src)
	// close(dst)
	// close(src)
}

func ExampleCopyFile_no_src() {
	dofakecopy(map[string]error{
		"open(src)": errors.New("src not found"),
	})
	// Output:
	// open(src)
	// copy(src, dst): src not found
}

func ExampleCopyFile_no_dst() {
	dofakecopy(map[string]error{
		"open(dst)": errors.New("dst not found"),
	})
	// Output:
	// open(src)
	// open(dst)
	// close(src)
	// copy(src, dst): dst not found
}

func ExampleCopyFile_copy_err() {
	dofakecopy(map[string]error{
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

func ExampleCopyFile_close_dst_err() {
	dofakecopy(map[string]error{
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

func dofakecopy(data map[string]error) {
	err := fakecopy(data)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
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
func fakecopy(data map[string]error) (err error) {
	check, done := handle.Errorf(&err, "copy(src, dst)")
	defer done()

	err = mock(data, "open(src)")
	check(err)
	defer mock(data, "close(src)")

	err = mock(data, "open(dst)")
	check(err)
	defer handle.Chain(&err, func() {
		mock(data, "close(dst)")
		mock(data, "remove(dst)")
	})

	err = mock(data, "copy(dst, src)")
	check(err)
	return mock(data, "close(dst)")
}

func mock(data map[string]error, name string) error {
	fmt.Printf("%s\n", name)
	return data[name]
}
