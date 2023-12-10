package builtins

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestPrintWorkingDirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() error
		wantW    string
		wantErr  bool
		wantErrW string
	}{
		{
			name: "success",
			setup: func() error {
				return nil
			},
			wantW: "/some/directory\n",
		},
		{
			name: "error",
			setup: func() error {
				return os.Chdir("/nonexistent/directory")
			},
			wantErr:  true,
			wantErrW: "failed to get current working directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.setup != nil {
				err := tt.setup()
				require.NoError(t, err, "setup failed")
				defer func() {
					err := os.Chdir("/")
					require.NoError(t, err, "cleanup failed")
				}()
			}

			w := &bytes.Buffer{}
			err := PrintWorkingDirectory(w)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrW)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantW, w.String())
			}
		})
	}
}
