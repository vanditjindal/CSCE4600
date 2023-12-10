package builtins

import (
	"fmt"
	"io"
	"os"
)

// AllocCommand defines the behavior of the 'alloc' shell builtin for csh/tcsh.

func AllocCommand(_ io.Writer, args ...string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: alloc VARNAME")
	}

	// Implement the allocation logic here.
	// For example, setting an environment variable:
	varName := args[1]
	value := "allocated_value"
	err := setEnvironmentVariable(varName, value)
	if err != nil {
		return err
	}

	return nil
}

// setEnvironmentVariable sets an environment variable.
func setEnvironmentVariable(name, value string) error {
	// Set the environment variable for the current process.
	err := os.Setenv(name, value)
	if err != nil {
		return fmt.Errorf("error setting environment variable: %v", err)
	}

	return nil
}
