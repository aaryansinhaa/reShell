# reshell

[![Go Version](https://img.shields.io/github/go-mod/go-version/aaryansinhaa/reshell?color=00ADD8)](https://go.dev/)
[![License](https://img.shields.io/github/license/aaryansinhaa/reshell?color=brightgreen)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen)](https://github.com/aaryansinhaa/reshell/actions)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-blueviolet.svg)](https://github.com/aaryansinhaa/reshell/pulls)

reshell is a portable developer environment and command-line workspace manager. It provides a terminal dashboard to configure, track, and synchronize aliases, script snippets, shell functions, environment variables, system packages, and git configurations from a single, version-controlled configuration directory.

---

## Tech Stack

- **Language**: Go (1.22+)
- **Terminal User Interface**: Charm Bubble Tea (MVU framework) & Lipgloss
- **Syntax Highlighting**: Chroma rendering engine
- **Supported Shells**: Bash, Zsh, Fish
- **Supported Package Managers**: APT, DNF, Pacman, Homebrew, Winget, Chocolatey

---

## System Architecture

```mermaid
graph TD
    User([User]) -->|Interact / Edit| TUI[reshell TUI Dashboard]
    TUI -->|Save Config| ConfigDir[~/.config/reshell/*.toml]
    CLI[reshell CLI / apply] -->|Read Config| ConfigDir
    CLI -->|Compile| OutScript[~/.config/reshell/shell/reshell.sh/fish]
    ShellProfile[Shell Profile: .bashrc / .zshrc / config.fish] -->|Sources on Startup| OutScript
    OutScript -->|Inject Context| ParentShell[Parent Shell Environment]
```

---

## Features

- **CLI Dashboard**: Terminal user interface for managing configurations.
- **Setup Wizard**: Detects installed text editors (Neovim, VS Code, Nano, etc.) to set editor preferences automatically.
- **Syntax Highlighting**: Real-time rendering for scripts, custom functions, and TOML templates.
- **Workspace Workflows**: Non-blocking, multi-step command and browser automation routines.
- **Cross-Platform Package Manager**: Asynchronously installs and uninstalls host packages on Linux, macOS, and Windows with secure sudo password piping.
- **Shell Compiler**: Generates optimized setup scripts and automatically registers startup hooks in `.bashrc`, `.zshrc`, or `config.fish`.
- **Portable Configurations**: Import or export configuration directories as a single TOML manifest. Configuration merges use a **Last-Write-Wins** policy, where matching incoming elements overwrite existing local settings.
- **Workspace Version Control**: Automatically tracks configuration files and custom scripts in a local Git repository, with TUI controls to view history and revert changes.

---

## Installation & Setup

### Prerequisites
- Go 1.22 or higher
- Git

### Build from Source
```bash
git clone https://github.com/aaryansinhaa/reshell.git
cd reshell
go build -o reshell
```

### Setup
Configure global binary path hooks and profile integrations:
```bash
./reshell setup
```
The setup command installs the `reshell` executable to `~/.local/bin/` and registers the startup hooks in your shell profile.

---

## Command Reference

| Command | Action |
| :--- | :--- |
| `reshell` | Launches the interactive TUI Dashboard |
| `reshell apply` | Compiles your active configurations and sources them |
| `reshell clean` | Removes reshell's integration blocks from your shell profile |
| `reshell setup` | Installs reshell binary globally and bootstraps configurations |
| `reshell alias add <name> <value> [desc]` | Registers a command alias |
| `reshell snippet add <name> <code> [desc]` | Stores a code block snippet |
| `reshell snippet copy <name>` | Copies snippet contents to your clipboard |
| `reshell function add <name> <code>` | Creates a custom shell script function |
| `reshell function validate <name>` | Runs a dry-run syntax diagnostic check |
| `reshell script run <cat> <name> [args]` | Runs a library script and writes output logs |
| `reshell workflow run <name>` | Runs a workflow sequence asynchronously |
| `reshell new <template> <name>` | Generates a project skeleton boilerplate |
| `reshell install [repo-url]` | Installs configuration packs or system packages |
| `reshell env add <name> <value>` | Registers environment variables |
| `reshell git apply` | Applies git profiles globally |
| `reshell git clear` | Clears version control history for the active profile |
| `reshell export <toml-path>` | Exports configurations into a single TOML manifest |
| `reshell import <toml-path>` | Imports configurations from a TOML manifest |
| `reshell profile list` | Lists all configuration profiles |
| `reshell profile create <name>` | Creates a new isolated configuration profile |
| `reshell profile switch <name>` | Switches active profile and recompiles hooks |
| `reshell profile delete <name>` | Deletes an isolated configuration profile |

---

## Configuration Architecture

All configurations are stored in your home directory under `~/.config/reshell/`:

```text
~/.config/reshell/
├── active_profile.txt  # Stores the currently active profile name
├── config.toml         # User info, preferred editor, packages, marketplace lists
├── aliases.toml        # Active command aliases
├── snippets.toml       # Script snippets & version history
├── env.toml            # Environment variables
├── workflows.toml      # Workflow definitions
├── functions/          # Raw custom function scripts (.sh, .fish)
├── scripts/            # Library scripts grouped by category
├── logs/               # Workflow and script execution logs
└── profiles/           # Isolated custom profile folders (e.g. school/, work/, chill/)
    └── school/
        ├── config.toml
        ├── aliases.toml
        └── ...
```

For comprehensive tutorials, setup guides, and marketplace documentation, refer to the [docs/](docs/) directory.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
