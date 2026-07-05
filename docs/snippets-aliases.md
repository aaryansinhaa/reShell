# Snippets & Aliases

This section explains how to manage script snippets and shell command mappings.

---

## Snippets

Snippets are reusable code blocks or templates stored in `~/.config/reshell/snippets.toml`.

<p align="center">
  <img src="../assets/reshell_snippets.jpg" alt="ReShell Snippets Management" width="650">
</p>

### Adding Snippets

Create a snippet using the command-line interface:

```bash
reshell snippet add <name> <code> [description]
```

Alternatively, open the dashboard (`reshell`), navigate to the **Snippets** tab, and press `n` to open the creation form.

### Revision History

When you update an existing snippet's code, reshell preserves the previous version. The historical script body is timestamped and stored in the `history` array inside `snippets.toml`:

```toml
[[snippets]]
name = "mkcd"
code = "mkdir -p \"$1\" && cd \"$1\""
description = "Make and change directory"

[[snippets.history]]
timestamp = "2026-06-28 17:00:00"
code = "mkdir -p \"$1\""
```

### Dashboard Actions
- **Editing (`e`)**: Opens the selected snippet code in a temporary `.txt` file using your preferred text editor (defined by `$EDITOR` or `config.toml`). Snippets are written as plain `.txt` files to prevent text editors from incorrectly forcing shell-script formatting and syntax highlighting on snippets written in other languages (such as C++ or Python).
- **Copying (`c`)**: Copies the highlighted snippet code directly to the host system clipboard.
- **Executing (`x`)**: Runs the snippet code in an isolated subshell process, printing stdout/stderr before returning to the dashboard.

---

## Aliases

Aliases map command shortcuts to longer terminal commands. They are stored in `~/.config/reshell/aliases.toml`.

<p align="center">
  <img src="../assets/reshell_aliases.jpg" alt="ReShell Aliases Management" width="650">
</p>

### Conflict Verification

When creating or modifying an alias, the engine verifies the name to prevent collision issues:

1. **System Path Check**: Verifies if the alias name collides with an existing binary in your `$PATH` (e.g., `ls` or `grep`).
2. **Function Verification**: Checks if a custom shell function is already registered with the same name.
3. **Duplicate Check**: Checks for duplicates in your existing alias list.

Warnings are displayed if conflicts are found, but you can override them if needed.

### Toggling State

To temporarily disable an alias, highlight it in the **Aliases** tab of the dashboard and press `Space`. Disabled aliases are excluded from the compiled configuration script during the next `reshell apply` execution.
