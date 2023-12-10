package builtins

import (
	"fmt"
	"io"
	"os"
)

// PrintWorkingDirectory prints the current working directory to the given writer.
func PrintWorkingDirectory(w io.Writer) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	_, err = fmt.Fprintln(w, wd)
	return err
}
