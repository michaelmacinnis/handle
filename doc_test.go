package handle

import (
	"errors"
	"fmt"

//	"github.com/michaelmacinnis/handle"
)

func do(name string) (err error) {
	// check, handle := handle.Errorf(&err, "do(%s)", name); defer handle()
	check, handle := Errorf(&err, "do(%s)", name); defer handle()

	// More compact that writing:
	//
	//     s, err = success()
	//     if err != nil {
	//         return err
	//     }
	//
	s, err := work(name)
	check(err)

	fmt.Printf("success(%s): %s\n", name, s)

	// The check can be cuddled to keep everything on one line.
	s, err = fail(name); check(err)

	// We will never reach here.
	fmt.Printf("failure(%s): %s\n", name, s)

	return nil
}

func fail(name string) (string, error) {
	return "", errors.New("failure")
}

func work(name string) (string, error) {
	return "Hello, " + name, nil
}

func Example() {
	do("World!")
	// Output: success(World!): Hello, World!
}
