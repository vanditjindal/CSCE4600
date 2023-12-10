package builtins

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddAlias(t *testing.T) {
	AliasMap.aliases = make(map[string]string) // Reset the alias map

	AddAlias("ll", "ls -l")
	AddAlias("gc", "git commit")

	assert.Equal(t, "ls -l", AliasMap.aliases["ll"])
	assert.Equal(t, "git commit", AliasMap.aliases["gc"])
}

func TestRemoveAlias(t *testing.T) {
	AliasMap.aliases = make(map[string]string) // Reset the alias map

	AddAlias("ll", "ls -l")
	AddAlias("gc", "git commit")

	RemoveAlias("ll")
	assert.NotContains(t, AliasMap.aliases, "ll")
	assert.Equal(t, "git commit", AliasMap.aliases["gc"])
}

func TestExecuteAlias(t *testing.T) {
	AliasMap.aliases = make(map[string]string) // Reset the alias map

	AddAlias("ll", "ls -l")
	AddAlias("gc", "git commit")

	w := &bytes.Buffer{}
	err := ExecuteAlias("ll", w)
	assert.NoError(t, err)
	assert.Contains(t, w.String(), "Executing command: ls -l")

	w.Reset()
	err = ExecuteAlias("echo Hello", w)
	assert.NoError(t, err)
	assert.Contains(t, w.String(), "Executing command: echo Hello")
}

func TestListAliases(t *testing.T) {
	AliasMap.aliases = make(map[string]string) // Reset the alias map

	AddAlias("ll", "ls -l")
	AddAlias("gc", "git commit")

	w := &bytes.Buffer{}
	ListAliases(w)

	expectedOutput := "Current aliases:\nll -> ls -l\ngc -> git commit\n"
	assert.Equal(t, expectedOutput, w.String())
}
