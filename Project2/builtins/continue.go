package builtins

import "fmt"

// OutputWriter is a function type for writing output (used for testing).
var OutputWriter func(string)

// ContinueShell simulates the behavior of the "continue" shell built-in.
func ContinueShell() {
	if OutputWriter != nil {
		OutputWriter("Continuing the shell...")
	} else {
		fmt.Println("Continuing the shell...")
	}
}
