package builtins_test

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"github.com/vanditjindal/CSCE4600/Project2/builtins" // Adjust this import path based on your project structure.
	"io"
	"os"
	"testing"
)

func TestAllocCommand(t *testing.T) {
	t.Parallel()

	type args struct {
		output io.Writer
		args   []string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		varName  string
		varValue string
	}{
		{
			name: "successful allocation",
			args: args{
				output: &bytes.Buffer{},
				args:   []string{"alloc", "VARNAME"},
			},
			varName:  "VARNAME",
			varValue: "allocated_value",
		},
		{
			name: "incorrect number of arguments",
			args: args{
				output: &bytes.Buffer{},
				args:   []string{"alloc"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Clear any existing environment variable for testing.
			_ = os.Unsetenv(tt.varName)

			// testing
			err := builtins.AllocCommand(tt.args.output, tt.args.args...)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// Check if the environment variable is set correctly.
				value := os.Getenv(tt.varName)
				require.Equal(t, tt.varValue, value, "unexpected environment variable value")
			}
		})
	}
}
