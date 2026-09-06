package printer

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
	fmt.Fprintf(out, "Use %s <command> --help\" for more information about a given command.\n", command.Name())
}

func PrintSubCommandHelp(out io.Writer, command *cobra.Command) {

	// 当前命令描述
	if command.Short != "" {
		fmt.Fprintln(out, command.Short)
		fmt.Fprintln(out)
	}

	// 如果存在子命令，先打印子命令分组
	if command.HasAvailableSubCommands() {
		firstGroup := true

		for _, group := range command.Groups() {
			commands := availableCommandsByGroup(command, group.ID)
			if len(commands) == 0 {
				continue
			}
			if !firstGroup {
				fmt.Fprintln(out)
			}
			fmt.Fprintf(out, "%s:\n", group.Title)
			for _, c := range commands {
				fmt.Fprintf(out, "  %-12s%s\n", c.Name(), c.Short)
			}
			firstGroup = false
		}

		// 子命令列表和Usage之间的空格
		fmt.Fprintln(out)
	}
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
