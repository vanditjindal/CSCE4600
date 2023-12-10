package main

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

func Test_runLoop(t *testing.T) {
	t.Parallel()
	exitCmd := strings.NewReader("exit\n")
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name     string
		args     args
		wantW    string
		wantErrW string
	}{
		{
			name: "no error",
			args: args{
				r: exitCmd,
			},
		},
		{
			name: "read error should have no effect",
			args: args{
				r: iotest.ErrReader(io.EOF),
			},
			wantErrW: "EOF",
		},
		{
			name: "pwd command",
			args: args{
				r: strings.NewReader("pwd\nexit\n"),
			},
			wantW: "/some/directory\n",
		},
		{
			name: "alias command",
			args: args{
				r: strings.NewReader("alias myalias=ls -l\nmyalias\nexit\n"),
			},
			wantW: "Executing alias: myalias -> ls\nCurrent aliases:\nmyalias -> ls\n",
		},
		{
			name: "alloc command",
			args: args{
				r: strings.NewReader("alloc VARNAME\nexit\n"),
			},
			wantW: "Setting environment variable VARNAME=allocated_value\n",
		},
		{
			name: "incorrect number of arguments",
			args: args{
				r: strings.NewReader("alloc\nexit\n"),
			},
			wantErrW: "usage: alloc VARNAME",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := &bytes.Buffer{}
			errW := &bytes.Buffer{}

			exit := make(chan struct{}, 2)
			// run the loop for 10ms
			go runLoop(tt.args.r, w, errW, exit)
			time.Sleep(10 * time.Millisecond)
			exit <- struct{}{}

			require.NotEmpty(t, w.String())
			if tt.wantErrW != "" {
				require.Contains(t, errW.String(), tt.wantErrW)
			} else {
				require.Empty(t, errW.String())
			}
		})
	}
	t.Run("continue shell", func(t *testing.T) {
		t.Parallel()
		continueCmd := strings.NewReader("continue\n")
		w := &bytes.Buffer{}
		errW := &bytes.Buffer{}

		exit := make(chan struct{}, 2)
		// run the loop for 10ms
		go runLoop(continueCmd, w, errW, exit)
		time.Sleep(10 * time.Millisecond)
		exit <- struct{}{}

		require.NotEmpty(t, w.String())
		require.Empty(t, errW.String())
	})
	t.Run("builtin command", func(t *testing.T) {
		t.Parallel()
		builtinCmd := strings.NewReader("builtin echo Hello, World!\nexit\n")
		w := &bytes.Buffer{}
		errW := &bytes.Buffer{}

		exit := make(chan struct{}, 2)
		// run the loop for 10ms
		go runLoop(builtinCmd, w, errW, exit)
		time.Sleep(10 * time.Millisecond)
		exit <- struct{}{}

		require.NotEmpty(t, w.String())
		require.Empty(t, errW.String())
		require.Equal(t, "Hello, World!\n", w.String())
	})
}
