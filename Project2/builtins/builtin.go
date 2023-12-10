package builtins

import (
	"fmt"
	"io"
)

// BuiltinCommand defines the behavior of the 'builtin' shell builtin.

func BuiltinCommand(output io.Writer, args ...string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: builtin COMMAND [ARGUMENTS...]")
	}

	// Execute the specified built-in command.
	switch args[1] {
	case "echo":
		// Replace with actual 'echo' implementation.
		_, _ = fmt.Fprintln(output, args[2:])
	default:
		return fmt.Errorf("unsupported built-in command: %s", args[1])
	}

	return nil
}
