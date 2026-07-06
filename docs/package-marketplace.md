# Package Installer & Marketplace

This section details how to manage system package dependencies and share configurations.

---

## Package Installer

The package installer automates host package installation across environments. Define the required packages in `config.toml`:

<p align="center">
  <img src="../assets/reshell_packages.jpg" alt="ReShell Package Management" width="650">
</p>

```toml
packages = [
    "git",
    "fzf",
    "bat",
    "ripgrep",
    "fd-find",
    "tmux",
    "lazygit"
]
```

### Dashboard Package Management
Navigate to the **Packages** tab in the dashboard to manage system requirements:

- **Add (`n`)**: Appends a package name to `config.toml`.
- **Remove (`d`)**: Removes a package name from the configuration list.
- **Uninstall (`u`)**: Asynchronously removes the highlighted package from your host system.
- **Install (`i`)**: Asynchronously installs all missing package dependencies.

### Privilege Elevation (sudo)

For package operations requiring administrative privileges (e.g., `apt-get`, `pacman`):

- reshell checks for active credentials using `sudo -n true` first. If cached, it bypasses the manual password prompt entirely.
- If not cached, reshell prompts you for your password inside the dashboard.
- The password is piped to `sudo -S` standard input to execute the installation asynchronously.
- The process streams command outputs directly to the dashboard log viewport.

---

## Marketplace Configuration Packs

Marketplace packages allow you to share environment configurations via Git repositories.

### Importing Packages

To import configurations:

```bash
reshell install github.com/aaryansinhaa/reshell-java
```

The import process:

1. Clones the remote Git repository into a temporary workspace.
2. Reads the `reshell.toml` manifest file from the repository root.
3. Displays a verification breakdown of all packages, variables, aliases, functions, and scripts, warns about third-party imports, and requests confirmation before merging.
4. If approved, merges the parsed aliases, snippets, and environment variables into your configuration.
5. Appends required packages to your global configuration list.
6. Copies custom functions in `functions/` to `~/.config/reshell/functions/`.
7. Copies scripts in `scripts/` to `~/.config/reshell/scripts/`.
8. Displays a post-install summary reporting all successfully modified assets.

### Manifest Schema (`reshell.toml`)

Example manifest for a marketplace configuration package:

```toml
[package]
name = "reshell-java"
description = "Java terminal configurations for developers"

[[aliases]]
name = "jrun"
value = "java -jar"
description = "Run a JAR file"
shell = "all"
enabled = true

[[variables]]
name = "JAVA_HOME"
value = "/usr/lib/jvm/java-17-openjdk-amd64"
description = "Java Home path"
enabled = true

[[snippets]]
name = "mvn-clean-install"
code = "mvn clean install -DskipTests"
description = "Maven build without running tests"
tags = ["maven", "java", "build"]
language = "bash"

[config]
packages = ["openjdk-17-jdk", "maven", "gradle"]
```
