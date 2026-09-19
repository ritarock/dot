# dot

## Usage
```bash
$ dot -h
A minimal copy-based dotfiles manager

The repository directory (default ~/dotfiles, override with $DOT_DIR)
mirrors the structure of $HOME, with leading dots renamed to "dot_"
(e.g. ~/.config/nvim -> dot_config/nvim). There is no config file.

Usage:
  dot [command]

Available Commands:
  add         copy files or directories from $HOME into the repo
  apply       copy repo -> $HOME (confirm unless -y)
  diff        show what apply would change
  help        Help about any command
  update      copy $HOME -> repo

Flags:
  -h, --help   help for dot

Use "dot [command] --help" for more information about a command.
```
