package handle_test

import (
	"fmt"

	"github.com/michaelmacinnis/handle"
)

func ExampleError_unmodified() {
	f := func() (err error) {
		check, handle := handle.Error(&err)
		defer handle()

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
	check, handle := handle.Error(&err, func() {
		fmt.Printf("err: %s\n", err.Error())
	})
	defer handle()

	s, err := works("World!")
	check(err)

	s, err = fails("World!")
	check(err)

	// We will never reach here
	fmt.Printf("%s\n", s)

	// Output: err: failure
}
