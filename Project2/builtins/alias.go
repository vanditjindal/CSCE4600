package builtins

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

// AliasMap represents the alias mapping.
var AliasMap = struct {
	sync.RWMutex
	aliases map[string]string
}{
	aliases: make(map[string]string),
}

// AddAlias adds an alias to the AliasMap.
func AddAlias(alias, command string) {
	AliasMap.Lock()
	defer AliasMap.Unlock()
	AliasMap.aliases[alias] = command
}

// RemoveAlias removes an alias from the AliasMap.
func RemoveAlias(alias string) {
	AliasMap.Lock()
	defer AliasMap.Unlock()
	delete(AliasMap.aliases, alias)
}

// ExecuteAlias executes a command by resolving aliases.
func ExecuteAlias(cmd string, output io.Writer) error {
	AliasMap.RLock()
	defer AliasMap.RUnlock()

	// Check if cmd is an alias.
	if alias, ok := AliasMap.aliases[cmd]; ok {
		// Execute the aliased command.
		_, _ = fmt.Fprintln(output, "Executing alias:", cmd, "->", alias)
		return executeCommand(alias)
	}

	// If not an alias, execute the original command.
	return executeCommand(cmd)
}

// ListAliases prints the list of aliases to the given writer.
func ListAliases(output io.Writer) {
	AliasMap.RLock()
	defer AliasMap.RUnlock()

	_, _ = fmt.Fprintln(output, "Current aliases:")
	for alias, cmd := range AliasMap.aliases {
		_, _ = fmt.Fprintf(output, "%s -> %s\n", alias, cmd)
	}
}

// AliasCommand defines the behavior of the 'alias' shell builtin.
func AliasCommand(output io.Writer, args ...string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: alias ALIAS_NAME=COMMAND")
	}

	aliasSplit := strings.SplitN(args[1], "=", 2)
	if len(aliasSplit) != 2 {
		return fmt.Errorf("invalid alias format: %s", args[1])
	}

	aliasName := aliasSplit[0]
	aliasCommand := aliasSplit[1]

	// Store the alias in the AliasMap.
	AddAlias(aliasName, aliasCommand)

	// Print feedback to the user.
	_, _ = fmt.Fprintf(output, "Executing alias: %s -> %s\n", aliasName, aliasCommand)
	ListAliases(output)

	return nil
}

// executeCommand is a mock function representing the actual command execution.
func executeCommand(cmd string) error {
	// In a real shell, you would replace this with the actual command execution logic.
	fmt.Println("Executing command:", cmd)
	return nil
}
