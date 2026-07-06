# reshell

[![Go Version](https://img.shields.io/github/go-mod/go-version/aaryansinhaa/reshell?color=00ADD8)](https://go.dev/)
[![License](https://img.shields.io/github/license/aaryansinhaa/reshell?color=brightgreen)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen)](https://github.com/aaryansinhaa/reshell/actions)
[![Coverage](https://img.shields.io/codecov/c/github/aaryansinhaa/reshell?logo=codecov)](https://codecov.io/gh/aaryansinhaa/reshell)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-blueviolet.svg)](https://github.com/aaryansinhaa/reshell/pulls)

reshell is a portable developer environment and command-line workspace manager. It provides a terminal dashboard to configure, track, and synchronize aliases, script snippets, shell functions, environment variables, system packages, and git configurations from a single, version-controlled configuration directory.

<p align="center">
  <img src="docs/assets/reshell_demo.jpg" alt="ReShell TUI Dashboard" width="800">
</p>

---

## Environment Bootstrapping

Setting up a new machine or server environment often requires hours of copying dotfiles, writing custom shell scripts, and installing dependencies individually. 

`reshell` automates this process into a single command. By importing pre-configured workspace packs from the marketplace or any Git repository, you can configure aliases, environment variables, custom helper scripts, and system package dependencies instantly.

### Example: Java Developer Workspace

To install a complete developer environment:

```bash
reshell install github.com/aaryansinhaa/reshell-java
```

reshell automatically:
1. Clones the remote repository and parses the `reshell.toml` manifest.
2. Merges development configurations:
   * **Aliases**: Binds shortcuts like `jrun` (`java -jar`).
   * **Environment Variables**: Sets paths like `JAVA_HOME`.
   * **Code Snippets**: Adds helper commands like `mvn-clean-install`.
   * **Custom Functions**: Copies shell functions to `~/.config/reshell/functions/`.
3. Appends required packages to the configuration manifest:
   ```toml
   packages = ["openjdk-17-jdk", "maven", "gradle"]
   ```

To install the missing dependencies and apply the configurations:

```bash
reshell install
```

`reshell` detects the host package manager (`apt`, `brew`, `dnf`, `pacman`, `winget`), prompts for administrative elevation if required, and installs the dependencies asynchronously.

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
- **Syntax Highlighting**: Real-time rendering for scripts, custom functions, and TOML templates.
- **Workspace Workflows**: Non-blocking, multi-step command and browser automation routines.
- **Cross-Platform Package Manager**: Asynchronously installs and uninstalls host packages on Linux, macOS, and Windows with secure sudo password piping.
- **Shell Compiler**: Generates optimized setup scripts and automatically registers startup hooks in `.bashrc`, `.zshrc`, or `config.fish`.
- **Portable Configurations**: Import or export configuration directories as a single TOML manifest. Note that the configuration merge overwrites the existing settings, but you can revert those using the version control available inside reshell.
- **Workspace Version Control**: Automatically tracks configuration files and custom scripts in a local Git repository, with TUI controls to view history and revert changes.

---

## Installation & Setup

### Run via Docker (Sandbox Demo)

For testing or running ReShell in an isolated sandbox without affecting your host system, you can use the pre-built Docker image hosted on GitHub Container Registry:

```bash
docker run -it ghcr.io/aaryansinhaa/reshell:latest
```

This starts an interactive Ubuntu container with `reshell` pre-installed, a default non-root user (`developer`), and common shell dependencies configured.

### Build from Source
```bash
git clone https://github.com/aaryansinhaa/reshell.git
cd reshell
go build -o reshell
```

### Setup
Configure global binary path hooks, profile integrations, and automatically import configurations:
```bash
./reshell setup [directory_path]
```
The setup command:

1. Installs the `reshell` executable to `~/.local/bin/` and registers startup hooks in your shell profile.
2. Prompts you to select or create a target configuration profile (default: `default`).
3. Auto-discovers and imports configurations:
   - If `directory_path` is provided, scans that folder recursively.
   - If no path is provided, scans default system shell profiles (`~/.bashrc`, `~/.zshrc`, `~/.profile`, etc.), recursively scans `~/.config` (skipping large folders), global VS Code user snippets, and Pet TOML configuration files.
4. Performs interactive de-duplication, conflict resolution (allowing you to choose between overriding, keeping existing, renaming, or skipping), and secure secret detection (skipped by default).

---

## Command Reference

| Command | Action |
| :--- | :--- |
| `reshell` | Launches the interactive TUI Dashboard |
| `reshell apply` | Compiles your active configurations and sources them |
| `reshell clean` | Removes reshell's integration blocks from your shell profile |
| `reshell setup [dir]` | Installs reshell binary globally and auto-imports dotfiles/configs |
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
├── snippets.toml       # User snippets
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
