package utils

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
)

func PrintRootHelp(out io.Writer, command *cobra.Command) {
	// Command groups
	firstGroup := true

	for _, group := range command.Groups() {
		commandsByGroup := availableCommandsByGroup(command, group.ID)
		if len(commandsByGroup) == 0 {
			continue
		}

		// Separate groups with one blank line.
		if !firstGroup {
			fmt.Fprintln(out)
		}
		fmt.Fprintf(out, "%s:\n", group.Title)

		for _, c := range commandsByGroup {
			fmt.Fprintf(
				out,
				"  %-12s%s\n",
				c.Name(),
				c.Short,
			)
		}
		firstGroup = false
	}
	//Usage
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")

	// Cobra automatically formats the flags:
	fmt.Fprintln(out, command.PersistentFlags().FlagUsages())

	// Help hint
	fmt.Fprintf(
		out,
		"Use %s <command> --help\" for more information about a given command.\n",
		command.Name(),
	)
}

func PrintCommandHelp(out io.Writer, command *cobra.Command) {
	// Usage
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintf(out, "  %s\n", command.UseLine())

	// Local flags
	if command.HasAvailableLocalFlags() {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Flags:")
		fmt.Fprint(out, command.LocalNonPersistentFlags().FlagUsages())
	}

	// Global flags
	if command.HasAvailableInheritedFlags() {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Global Flags:")
		fmt.Fprint(out, command.InheritedFlags().FlagUsages())
	}

	// Subcommands
	if command.HasAvailableSubCommands() {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Subcommands:")
		for _, cmd := range command.Commands() {
			if !cmd.IsAvailableCommand() {
				continue
			}
			fmt.Fprintf(out, "  %-12s%s\n", cmd.Name(), cmd.Short)
		}
		fmt.Fprintln(out)
		fmt.Fprintf(
			out,
			`Use "%s <command> --help" for more information about a given command.`,
			command.CommandPath(),
		)
		fmt.Fprintln(out)
	}
}

func availableCommandsByGroup(command *cobra.Command, groupid string) []*cobra.Command {
	var commands []*cobra.Command
	for _, cmd := range command.Commands() {
		if !cmd.IsAvailableCommand() {
			continue
		}
		if cmd.GroupID != groupid {
			continue
		}
		commands = append(commands, cmd)
	}
	return commands
}

// PrintUsage 统一打印命令使用提示
func PrintUsage(cmd *cobra.Command, msg string) {
	cmd.Help()
	fmt.Println()
	if msg != "" {
		fmt.Println(msg)
	}
}
