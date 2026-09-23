package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [task]...",
	Short: "run tasks defined in runner.yaml (confirm unless -y)",
	Long: `run tasks defined in runner.yaml (confirm unless -y)

Tasks are read from runner.yaml at the repository root. It is the one file
in the repository that is never copied to $HOME.

  tasks:
    - name: brew
      run: brew install jq ripgrep fzf
    - name: mise
      run: mise install

With no arguments every task runs; with arguments only the named tasks run,
in the given order. Each "run" is executed with "sh -c" from the repository
root, inheriting the current environment. Execution stops at the first task
that exits non-zero.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _, err := dirs()
		if err != nil {
			return err
		}
		yes, _ := cmd.Flags().GetBool("yes")
		return cmdRun(repo, args, yes, cmd.InOrStdin())
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolP("yes", "y", false, "skip confirmation")
}

func cmdRun(repo string, names []string, yes bool, stdin io.Reader) error {
	tasks, err := loadTasks(repo)
	if err != nil {
		return err
	}

	tasks, err = selectTasks(tasks, names)
	if err != nil {
		return err
	}

	return runTasks(tasks, repo, yes, stdin, execTask)
}

func runTasks(tasks []task, dir string, yes bool, stdin io.Reader, run runFunc) error {
	if !yes {
		for _, t := range tasks {
			fmt.Printf("%s: %s\n", t.Name, t.Run)
		}
		if !confirm(fmt.Sprintf("Run %d task(s)? [y/N]: ", len(tasks)), stdin) {
			fmt.Println("aborted")
			return nil
		}
	}

	for i, t := range tasks {
		if i > 0 {
			fmt.Println()
		}
		fmt.Println("==>", t.Name)
		if err := run(t, dir); err != nil {
			return fmt.Errorf("task %q failed: %v", t.Name, err)
		}
	}

	return nil
}
