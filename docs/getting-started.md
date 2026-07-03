# Getting Started

This guide covers requirements, installation, shell profile integration, and import/export commands.

## Requirements

- **Go**: Version 1.22 or higher.
- **Operating Systems**: Linux, macOS, or Windows (via PowerShell or Git Bash).
- **Git**: Required for configuration version control and importing marketplace packs.

---

## Installation & Setup

Build the binary from the repository root:

```bash
go build -o reshell
```

Initialize the global configuration, install the binary, and inject startup hooks:

```bash
./reshell setup
```

The `setup` command automatically copy-installs the `reshell` executable into your local user path (`~/.local/bin/`), registers the path variables, and injects the shell hook integrations. Open a new terminal window to begin using `reshell` globally.

---

## Active Shell Integration

reshell compiles your configurations and hooks them into your startup profile. Run:

```bash
reshell apply
```

This generates shell-specific script outputs and registers startup hooks:
- **Zsh**: Adds integration blocks to `~/.zshrc` and compiles imports to `~/.config/reshell/shell/reshell.sh`.
- **Bash**: Adds integration blocks to `~/.bashrc` and compiles imports to `~/.config/reshell/shell/reshell.sh`.
- **Fish**: Adds integration blocks to `~/.config/fish/config.fish` and compiles imports to `~/.config/reshell/shell/reshell.fish`.

### Startup Hook Structure

The injected configuration block in your shell profile:

```bash
# >>> reshell initialize >>>
if [ -f "$HOME/.config/reshell/shell/reshell.sh" ]; then
    . "$HOME/.config/reshell/shell/reshell.sh"
fi
# <<< reshell initialize <<<
```

To clean all reshell integration blocks and restore your profile files, run:

```bash
reshell clean
```

---

## Configuration Portability (Import & Export)

To prevent configuration drift, you can export and import your workspace configuration as a unified TOML manifest or back up the raw config folder.

### Conflict Resolution (Last-Write-Wins Policy)

reshell uses a **Last-Write-Wins (LWW)** conflict resolution policy when merging configurations:
- **Manifest Imports (`reshell import`)**: Overwrites the existing local configurations with the contents of the imported TOML manifest file.
- **Marketplace Packs (`reshell install`)**: Merges configurations item-by-item. If an imported alias, environment variable, or snippet matches an existing local key, the imported value overwrites the local one.
- **Git Version Control**: Since reshell automatically commits all changes under `~/.config/reshell/`, you can inspect diffs and resolve conflicts or revert undesired overwrites using standard Git command-line tools.

### Exporting Configurations

To export environment variables, aliases, snippets, package lists, and workflows into a single TOML manifest:

```bash
reshell export ~/backup-config.toml
```

### Importing Configurations

To import configurations from a manifest and merge them with your current setup:

```bash
reshell import ~/backup-config.toml
```

Once imported, execute `reshell apply` to compile and source the new configuration.

---

## Dashboard Usage

To launch the interactive configuration editor, run the binary without any subcommands:

```bash
reshell
```

### Keyboard Shortcuts
- `Tab`: Navigate forward through sidebar tabs.
- `Shift+Tab`: Navigate backward through sidebar tabs.
- `Up / Down` (or `k / j`): Scroll item lists.
- `n`: Create a new entry (opens input form).
- `e`: Open the selected custom function or script in your default editor (`$EDITOR`).
- `d`: Delete the selected entry.
- `Space`: Toggle the active state of an environment variable or alias.
- `c`: Copy the selected script snippet to the system clipboard.
- `x`: Execute the selected script or workflow.
- `h` (inside Git tab): Toggle between global git configuration and local repository version history.
- `r` or `Enter` (inside Git history view): Revert configuration files to the selected revision.
- `Ctrl+A`: Run `reshell apply` to compile and load configurations.
- `q` or `Ctrl+C`: Exit the interface.
