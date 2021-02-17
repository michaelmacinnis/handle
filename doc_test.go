package handle_test

import (
	"errors"
	"fmt"

	"github.com/michaelmacinnis/handle"
)

func do(name string) (err error) {
	check, handle := handle.Errorf(&err, "do(%s)", name)
	defer handle()

	// More compact than writing:
	//
	//     s, err = success()
	//     if err != nil {
	//         return fmt.Errorf("do(%s): %w", name, err)
	//     }
	//
	s, err := works(name)
	check(err)

	fmt.Printf("works(%s): %s\n", name, s)

	// The check can be cuddled to keep everything on one line.
	s, err = fails(name); check(err)

	// We will never reach here.
	fmt.Printf("fails(%s): %s\n", name, s)

	return nil
}

func fails(name string) (string, error) {
	return "", errors.New("failure")
}

func works(name string) (string, error) {
	return "Hello, " + name, nil
}

func Example() {
	do("World!")
	// Output: works(World!): Hello, World!
}
