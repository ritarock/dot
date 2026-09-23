# dot

## Usage
```bash
$ dot -h
A minimal copy-based dotfiles manager

The repository directory (default ~/dotfiles, override with $DOT_DIR)
mirrors the structure of $HOME, with leading dots renamed to "dot_"
(e.g. ~/.config/nvim -> dot_config/nvim). Copying has no configuration;
the only special file is an optional runner.yaml at the repository root,
which defines the tasks that "dot run" executes.

Usage:
  dot [command]

Available Commands:
  add         copy files or directories from $HOME into the repo
  apply       copy repo -> $HOME (confirm unless -y)
  diff        show what apply would change
  help        Help about any command
  run         run tasks defined in runner.yaml (confirm unless -y)
  update      copy $HOME -> repo

Flags:
  -h, --help   help for dot

Use "dot [command] --help" for more information about a command.
```

```bash
$ dot run -h
run tasks defined in runner.yaml (confirm unless -y)

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
that exits non-zero.

Usage:
  dot run [task]... [flags]

Flags:
  -h, --help   help for run
  -y, --yes    skip confirmation
```
