package builtins_test

import (
	"testing"

	"github.com/vanditjindal/CSCE4600/Project2/builtins"
)

func TestContinueShell(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "continue shell",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the output to check if fmt.Println is called
			var output string
			builtins.OutputWriter = func(s string) {
				output = s
			}

			builtins.ContinueShell()

			// Check if fmt.Println is called with the expected message
			expectedOutput := "Continuing the shell..."
			if output != expectedOutput {
				t.Errorf("ContinueShell() output = %v, want %v", output, expectedOutput)
			}
		})
	}
}
