package packages

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"reshell/pkg/config"
	"strings"
)

// DetectOS returns the OS type and default package manager name.
func DetectOS() (string, string) {
	// Simple platforms
	if os.Getenv("OS") == "Windows_NT" {
		if _, err := exec.LookPath("winget"); err == nil {
			return "windows", "winget"
		}
		return "windows", "choco"
	}

	// Read /etc/os-release for Linux distros
	if _, err := os.Stat("/etc/os-release"); err == nil {
		data, err := os.ReadFile("/etc/os-release")
		if err == nil {
			content := string(data)
			if strings.Contains(content, "ubuntu") || strings.Contains(content, "debian") {
				return "linux", "apt"
			}
			if strings.Contains(content, "fedora") || strings.Contains(content, "centos") || strings.Contains(content, "rhel") {
				return "linux", "dnf"
			}
			if strings.Contains(content, "arch") || strings.Contains(content, "manjaro") {
				return "linux", "pacman"
			}
		}
	}

	// Check macOS
	if _, err := exec.LookPath("brew"); err == nil {
		return "darwin", "brew"
	}

	// fallback Linux
	return "linux", "apt"
}

// IsInstalled checks if a package/binary is installed on the system.
func IsInstalled(pkgName string) bool {
	// 1. Check path lookup for common binaries
	binaryName := pkgName
	switch pkgName {
	case "ripgrep":
		binaryName = "rg"
	case "fd":
		binaryName = "fd"
		// On Ubuntu, fd-find package installs fdfind binary
		if _, err := exec.LookPath("fdfind"); err == nil {
			return true
		}
	case "bat":
		binaryName = "bat"
		// On Ubuntu, bat package installs batcat binary
		if _, err := exec.LookPath("batcat"); err == nil {
			return true
		}
	}

	if _, err := exec.LookPath(binaryName); err == nil {
		return true
	}

	// 2. Query package manager if binary lookup fails
	_, manager := DetectOS()
	var cmd *exec.Cmd

	switch manager {
	case "apt":
		cmd = exec.Command("dpkg", "-s", pkgName)
	case "dnf":
		cmd = exec.Command("rpm", "-q", pkgName)
	case "pacman":
		cmd = exec.Command("pacman", "-Qq", pkgName)
	case "brew":
		cmd = exec.Command("brew", "list", pkgName)
	default:
		return false
	}

	err := cmd.Run()
	return err == nil
}

// Install runs the OS package installer asynchronously, writing console feedback to stdoutChan.
func Install(pkgName, manager string, sudoPassword []byte, stdoutChan chan<- string) error {
	if len(sudoPassword) > 0 {
		defer func() {
			for i := range sudoPassword {
				sudoPassword[i] = 0
			}
		}()
	}

	var cmd *exec.Cmd

	// Setup manager commands
	switch manager {
	case "apt":
		if len(sudoPassword) > 0 {
			cmd = exec.Command("sudo", "-S", "apt-get", "install", "-y", pkgName)
		} else {
			cmd = exec.Command("apt-get", "install", "-y", pkgName)
		}
	case "dnf":
		if len(sudoPassword) > 0 {
			cmd = exec.Command("sudo", "-S", "dnf", "install", "-y", pkgName)
		} else {
			cmd = exec.Command("dnf", "install", "-y", pkgName)
		}
	case "pacman":
		if len(sudoPassword) > 0 {
			cmd = exec.Command("sudo", "-S", "pacman", "-S", "--noconfirm", pkgName)
		} else {
			cmd = exec.Command("pacman", "-S", "--noconfirm", pkgName)
		}
	case "brew":
		cmd = exec.Command("brew", "install", pkgName)
	case "winget":
		cmd = exec.Command("winget", "install", "-e", "--id", pkgName, "--silent")
	case "choco":
		cmd = exec.Command("choco", "install", "-y", pkgName)
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}

	// Set up stdin piping for sudo
	var stdinPipe io.WriteCloser
	if len(sudoPassword) > 0 && (manager == "apt" || manager == "dnf" || manager == "pacman") {
		var err error
		stdinPipe, err = cmd.StdinPipe()
		if err != nil {
			return err
		}
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Feed sudo password if piped
	if stdinPipe != nil {
		_, _ = stdinPipe.Write(sudoPassword)
		_, _ = stdinPipe.Write([]byte("\n"))
		stdinPipe.Close()
	}

	// Stream stdout & stderr outputs
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				stdoutChan <- string(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				stdoutChan <- string(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	return cmd.Wait()
}

// Uninstall runs the OS package uninstaller asynchronously, writing console feedback to stdoutChan.
func Uninstall(pkgName, manager string, sudoPassword []byte, stdoutChan chan<- string) error {
	if len(sudoPassword) > 0 {
		defer func() {
			for i := range sudoPassword {
				sudoPassword[i] = 0
			}
		}()
	}

	var cmd *exec.Cmd

	switch manager {
	case "apt":
		if len(sudoPassword) > 0 {
			cmd = exec.Command("sudo", "-S", "apt-get", "remove", "-y", pkgName)
		} else {
			cmd = exec.Command("apt-get", "remove", "-y", pkgName)
		}
	case "dnf":
		if len(sudoPassword) > 0 {
			cmd = exec.Command("sudo", "-S", "dnf", "remove", "-y", pkgName)
		} else {
			cmd = exec.Command("dnf", "remove", "-y", pkgName)
		}
	case "pacman":
		if len(sudoPassword) > 0 {
			cmd = exec.Command("sudo", "-S", "pacman", "-R", "--noconfirm", pkgName)
		} else {
			cmd = exec.Command("pacman", "-R", "--noconfirm", pkgName)
		}
	case "brew":
		cmd = exec.Command("brew", "uninstall", pkgName)
	case "winget":
		cmd = exec.Command("winget", "uninstall", pkgName, "--silent")
	case "choco":
		cmd = exec.Command("choco", "uninstall", "-y", pkgName)
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}

	// Set up stdin piping for sudo
	var stdinPipe io.WriteCloser
	if len(sudoPassword) > 0 && (manager == "apt" || manager == "dnf" || manager == "pacman") {
		var err error
		stdinPipe, err = cmd.StdinPipe()
		if err != nil {
			return err
		}
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Feed sudo password if piped
	if stdinPipe != nil {
		_, _ = stdinPipe.Write(sudoPassword)
		_, _ = stdinPipe.Write([]byte("\n"))
		stdinPipe.Close()
	}

	// Stream stdout & stderr outputs
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				stdoutChan <- string(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				stdoutChan <- string(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	return cmd.Wait()
}

// Add appends a new package name to config.toml.
func Add(pkgName string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	for _, p := range cfg.Packages {
		if p == pkgName {
			return nil
		}
	}
	cfg.Packages = append(cfg.Packages, pkgName)
	return config.SaveConfig(cfg)
}

// Remove deletes a package name from config.toml.
func Remove(pkgName string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	newPkgs := make([]string, 0, len(cfg.Packages))
	for _, p := range cfg.Packages {
		if p != pkgName {
			newPkgs = append(newPkgs, p)
		}
	}
	cfg.Packages = newPkgs
	return config.SaveConfig(cfg)
}
