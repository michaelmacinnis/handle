package handle_test

import (
	"errors"
	"fmt"

	"github.com/michaelmacinnis/handle"
)

func do(name string) (err error) {
	check, handle := handle.Errorf(&err, "do(%s)", name); defer handle()

	// More compact that writing:
	//
	//     s, err = success()
	//     if err != nil {
	//         return err
	//     }
	//
	s, err := success(name)
	check(err)

	fmt.Printf("success(%s): %s\n", name, s)

	// The check can be cuddled to keep everything on one line.
	s, err = failure(name); check(err)

	// We will never reach here.
	fmt.Printf("failure(%s): %s\n", name, s)

	return nil
}

func failure(name string) (string, error) {
	return "", errors.New("failure")
}

func success(name string) (string, error) {
	return "Hello, " + name, nil
}

func Example() {
	do("World!")
	// Output: success(World!): Hello, World!
}
