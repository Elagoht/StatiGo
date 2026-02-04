// Package cli provides command-line interface utilities for the Statigo framework.
package cli

import (
	"fmt"
	"os"
)

// Command represents a CLI command.
type Command struct {
	Name    string
	Aliases []string
	Desc    string
	Run     func() error
}

// CLI manages command-line interface.
type CLI struct {
	commands map[string]*Command
}

// New creates a new CLI instance.
func New() *CLI {
	return &CLI{
		commands: make(map[string]*Command),
	}
}

// Register registers a command with its aliases.
func (c *CLI) Register(cmd *Command) {
	// Register primary name
	c.commands[cmd.Name] = cmd

	// Register all aliases
	for _, alias := range cmd.Aliases {
		c.commands[alias] = cmd
	}
}

// Execute runs a command by name.
func (c *CLI) Execute(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	cmdName := args[0]
	cmd, exists := c.commands[cmdName]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	return cmd.Run()
}

// PrintHelp prints available commands.
func (c *CLI) PrintHelp() {
	fmt.Println("Available commands:")

	// Track which commands we've printed to avoid duplicates from aliases
	printed := make(map[string]bool)

	for name, cmd := range c.commands {
		if name == cmd.Name && !printed[cmd.Name] {
			aliasStr := ""
			if len(cmd.Aliases) > 0 {
				aliasStr = fmt.Sprintf(" (aliases: %v)", cmd.Aliases)
			}
			fmt.Printf("  %s%s - %s\n", cmd.Name, aliasStr, cmd.Desc)
			printed[cmd.Name] = true
		}
	}
}

// ParseCommand parses command line arguments and determines if a command is being run.
func ParseCommand(args []string) (string, bool) {
	if len(args) < 2 {
		return "", false
	}

	return args[1], true
}

// ShouldRunCommand checks if we should run a command instead of starting the server.
func ShouldRunCommand() bool {
	cmd, exists := ParseCommand(os.Args)
	if !exists {
		return false
	}

	// List of known commands and aliases
	knownCommands := map[string]bool{
		"prerender":   true,
		"pre-render":  true,
		"bake":        true,
		"warm":        true,
		"prepare":     true,
		"cache-all":   true,
		"clear-cache": true,
		"invalidate":  true,
	}

	return knownCommands[cmd]
}
