package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

func A() error {
	return ErrNotFound
}

func B() error {
	err := A()
	return fmt.Errorf("b: %w", err)
}

// *Hint: look at errors.Is
// as a followup, read about errors.As and see if you can use it or thing cases where it might be useful

func main() {
	err := B()
	// TODO determine if err var is an ErrNotFound
	if errors.Is(err, ErrNotFound) {
		fmt.Println("Is ErrNotFound")
	}

	if errors.As(err, &ErrNotFound) {
		fmt.Println("As ErrNotFound")
	}
}
