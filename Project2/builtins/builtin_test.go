package builtins_test

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"github.com/vanditjindal/CSCE4600/Project2/builtins"
	"io"
	"testing"
)

func TestBuiltinCommand(t *testing.T) {
	t.Parallel()

	type args struct {
		output io.Writer
		args   []string
	}
	tests := []struct {
		name     string
		args     args
		wantW    string
		wantErrW string
	}{
		{
			name: "echo command",
			args: args{
				output: &bytes.Buffer{},
				args:   []string{"builtin", "echo", "Hello, World!"},
			},
			wantW: "Hello, World!\n",
		},
		{
			name: "unsupported command",
			args: args{
				output: &bytes.Buffer{},
				args:   []string{"builtin", "unsupported"},
			},
			wantErrW: "unsupported built-in command: unsupported",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// testing
			err := builtins.BuiltinCommand(tt.args.output, tt.args.args...)
			if tt.wantErrW != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrW)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantW, tt.args.output.(*bytes.Buffer).String())
			}
		})
	}
}
